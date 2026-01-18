use anyhow::Result;
use crossterm::{
    event::{self, Event, KeyCode, KeyEventKind},
    execute,
    terminal::{disable_raw_mode, enable_raw_mode, EnterAlternateScreen, LeaveAlternateScreen},
};
use futures::StreamExt;
use ratatui::{
    prelude::*,
    widgets::{Block, BorderType, Borders, Cell, Clear, Paragraph, Row, Table, TableState, Dataset, Chart, Axis, GraphType},
    symbols,
};
use serde::{Deserialize, Serialize};
use std::io;
use std::process::Stdio;
use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant};
use tokio::io::{AsyncBufReadExt, BufReader};
use tokio::process::Command;

// --- Palette ---
const COLOR_ACCENT: Color = Color::Rgb(255, 30, 30); // Red Dark Neon
const COLOR_PRIMARY: Color = Color::Rgb(189, 147, 249); // Purple
const COLOR_DIM: Color = Color::DarkGray;
const COLOR_TEXT: Color = Color::White;
const COLOR_WARN: Color = Color::Yellow;
const COLOR_SUCCESS: Color = Color::Green;
const COLOR_ERROR: Color = Color::Red;

// --- Events ---
#[derive(Serialize, Deserialize, Debug, Clone)]
struct BackendEvent {
    #[serde(rename = "type")]
    kind: Option<String>,
    message: Option<String>,
    level: Option<String>,
    msg: Option<String>,
}

#[derive(PartialEq, Clone, Copy, Debug)]
enum CurrentScreen {
    Login, // Start here
    Table,
    Monitor,
    UrlInput, 
}

#[derive(PartialEq, Clone, Copy, Debug)]
enum LoginField {
    Username,
    Password,
}

struct App {
    screen: CurrentScreen,
    logs: Arc<Mutex<Vec<String>>>,
    
    // Login State
    login_field: LoginField,
    username_input: String,
    password_input: String,
    login_error: String,
    
    // Scanner
    scan_active: bool,
    scan_progress: f64,
    scan_status: String,
    url_input: String, 
    
    // Monitor Data
    cpu_data: Vec<(f64, f64)>,
    mem_data: Vec<(f64, f64)>,
    window: [f64; 2],

    // Table
    table_state: TableState,
    items: Vec<(String, String, String, String)>,
    start_time: Instant,
    should_quit: bool,
}

impl App {
    fn new() -> App {
        let mut table_state = TableState::default();
        table_state.select(Some(0));

        App {
            screen: CurrentScreen::Login,
            logs: Arc::new(Mutex::new(vec!["Authentication Required.".to_string()])),
            
            login_field: LoginField::Username,
            username_input: String::new(),
            password_input: String::new(),
            login_error: String::new(),

            scan_active: false,
            scan_progress: 0.0,
            scan_status: "LOCKED".to_string(), // Locked status
            url_input: "https://".to_string(), 
            cpu_data: vec![],
            mem_data: vec![],
            window: [0.0, 100.0],
            table_state,
            items: vec![
                ("scan-target".to_string(), "Security Job".to_string(), "Ready".to_string(), "Start Phishing Scan".to_string()),
                ("infra-up".to_string(), "Infrastructure".to_string(), "Stopped".to_string(), "docker-compose up -d".to_string()),
                ("infra-down".to_string(), "Infrastructure".to_string(), "-".to_string(), "docker-compose down".to_string()),
                ("view-graph".to_string(), "Visual AI".to_string(), "Online".to_string(), "open neo4j browser".to_string()),
                ("check-storage".to_string(), "Storage".to_string(), "Online".to_string(), "open minio console".to_string()),
                ("check-queue".to_string(), "Messaging".to_string(), "Active".to_string(), "open rabbitmq".to_string()),
                ("run-tests".to_string(), "System".to_string(), "-".to_string(), "go test ./...".to_string()),
            ],
            start_time: Instant::now(),
            should_quit: false,
        }
    }
    
    fn on_tick(&mut self) {
        if self.screen == CurrentScreen::Login { return; } // Pause ticks on login
        
        let now = self.start_time.elapsed().as_secs_f64();
        let signal = if self.scan_active { (now * 5.0).sin() * 20.0 + 60.0 } else { (now * 0.5).sin() * 5.0 + 10.0 };
        self.cpu_data.push((now, signal.abs()));
        self.mem_data.push((now, signal.abs() * 0.8 + 20.0));
        if self.cpu_data.len() > 100 { self.cpu_data.remove(0); }
        if self.mem_data.len() > 100 { self.mem_data.remove(0); }
        self.window = [now - 50.0, now];
        if self.window[0] < 0.0 { self.window[0] = 0.0; }
        if self.scan_active && self.scan_progress < 1.0 { self.scan_progress += 0.005; }
    }

    fn next_row(&mut self) {
        let i = match self.table_state.selected() {
            Some(i) => if i >= self.items.len() - 1 { 0 } else { i + 1 },
            None => 0,
        };
        self.table_state.select(Some(i));
    }

    fn prev_row(&mut self) {
        let i = match self.table_state.selected() {
            Some(i) => if i == 0 { self.items.len() - 1 } else { i - 1 },
            None => 0,
        };
        self.table_state.select(Some(i));
    }
    
    fn action_run_browser(&mut self, url: &str) {
         self.logs.lock().unwrap().push(format!("[CMD] Opening Browser: {}", url));
         let _ = std::process::Command::new("cmd").args(&["/C", "start", url]).spawn();
    }
    
    fn action_run_docker(&mut self, args: &[&str]) {
         let args_vec: Vec<String> = args.iter().map(|s| s.to_string()).collect();
         let logs_ref = self.logs.clone();
         
         self.logs.lock().unwrap().push(format!("[CMD] Running Docker: {:?}", args));
         
         tokio::spawn(async move {
             run_docker_async(logs_ref, args_vec).await;
         });
    }
}

async fn run_docker_async(logs: Arc<Mutex<Vec<String>>>, args: Vec<String>) {
    // Robust Path Finding: Find where 'deploy/docker-compose.yml' actually is
    // This handles running from 'cli-rust', 'target/debug', or root.
    let mut current_dir = std::env::current_dir().unwrap_or_else(|_| std::path::PathBuf::from("."));
    let mut project_root = current_dir.clone();
    let mut found = false;
    
    // Traverse up to 5 levels to find 'deploy' folder
    for _ in 0..5 {
        if project_root.join("deploy").join("docker-compose.yml").exists() {
            found = true;
            break;
        }
        if !project_root.pop() { break; }
    }
    
    // Fallback if not found: Assume ../ from default behavior
    if !found {
        logs.lock().unwrap().push("[WARN] Could not auto-detect project root. using default '..' path.".to_string());
        project_root = std::path::PathBuf::from(".."); 
    } else {
         logs.lock().unwrap().push(format!("[DEBUG] Project Root Found: {:?}", project_root));
    }

    let mut child = Command::new("docker")
        .args(&args)
        .current_dir(project_root) // Use the robustly found root
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn();

    match child {
        Ok(mut child) => {
             // Capture stdout
             if let Some(stdout) = child.stdout.take() {
                 let logs = logs.clone();
                 tokio::spawn(async move {
                     let mut reader = BufReader::new(stdout).lines();
                     while let Ok(Some(line)) = reader.next_line().await {
                         logs.lock().unwrap().push(format!("[DOCKER] {}", line));
                     }
                 });
             }
             // Capture stderr (Docker Compose uses this for progress)
             if let Some(stderr) = child.stderr.take() {
                 let logs = logs.clone();
                 tokio::spawn(async move {
                     let mut reader = BufReader::new(stderr).lines();
                     while let Ok(Some(line)) = reader.next_line().await {
                         logs.lock().unwrap().push(format!("[DOCKER] {}", line));
                     }
                 });
             }
             
             let _ = child.wait().await;
             logs.lock().unwrap().push(format!("[DOCKER] Command Finished: {:?}", args));
        }
        Err(e) => {
             logs.lock().unwrap().push(format!("[ERROR] Docker failed: {}", e));
        }
    }
}

// --- Main Loop ---
#[tokio::main]
async fn main() -> Result<()> {
    enable_raw_mode()?;
    let mut stdout = io::stdout();
    execute!(stdout, EnterAlternateScreen)?;
    let backend = CrosstermBackend::new(stdout);
    let mut terminal = Terminal::new(backend)?;

    let app = Arc::new(Mutex::new(App::new()));
    let app_clone = app.clone();

    // Event Loop
    let _input_handle = tokio::spawn(async move {
        let mut tick_interval = tokio::time::interval(Duration::from_millis(100));
        let mut reader = crossterm::event::EventStream::new();

        loop {
            tokio::select! {
                _ = tick_interval.tick() => {
                    let mut app = app_clone.lock().unwrap();
                    app.on_tick();
                }
                Some(Ok(event)) = reader.next() => {
                    if let Event::Key(key) = event {
                        if key.kind == KeyEventKind::Press {
                            let mut app = app_clone.lock().unwrap();
                            
                            match app.screen {
                                CurrentScreen::Login => {
                                    match key.code {
                                        KeyCode::Esc => { app.should_quit = true; return; },
                                        KeyCode::Tab => {
                                            app.login_field = match app.login_field {
                                                LoginField::Username => LoginField::Password,
                                                LoginField::Password => LoginField::Username,
                                            };
                                        },
                                        KeyCode::Enter => {
                                            match app.login_field {
                                                LoginField::Username => {
                                                    // Move focus to password
                                                    app.login_field = LoginField::Password;
                                                },
                                                LoginField::Password => {
                                                    // Submit logic
                                                    if app.username_input == "phishvault" && app.password_input == "phishvault-tp2" {
                                                        app.screen = CurrentScreen::Table;
                                                        app.scan_status = "IDLE".to_string();
                                                        app.logs.lock().unwrap().push("Admin Authenticated.".to_string());
                                                        app.login_error.clear();
                                                    } else {
                                                        app.login_error = "Invalid Credentials!".to_string();
                                                        app.password_input.clear();
                                                        // Reset focus to try again? Optional. Keep at password.
                                                    }
                                                }
                                            }
                                        },
                                        KeyCode::Char(c) => {
                                            match app.login_field {
                                                LoginField::Username => app.username_input.push(c),
                                                LoginField::Password => app.password_input.push(c),
                                            }
                                        },
                                        KeyCode::Backspace => {
                                            match app.login_field {
                                                LoginField::Username => { app.username_input.pop(); },
                                                LoginField::Password => { app.password_input.pop(); },
                                            }
                                        },
                                        _ => {}
                                    }
                                },
                                CurrentScreen::Table => {
                                    // Global Quit only in main screens
                                    if key.code == KeyCode::Char('q') { app.should_quit = true; return; }
                                    
                                    match key.code {
                                        KeyCode::Char('j') | KeyCode::Down => app.next_row(),
                                        KeyCode::Char('k') | KeyCode::Up => app.prev_row(),
                                        KeyCode::Tab => app.screen = CurrentScreen::Monitor,
                                        KeyCode::Enter => {
                                            let selected = app.table_state.selected().unwrap_or(0);
                                            match app.items[selected].0.as_str() {
                                                "scan-target" => {
                                                    app.screen = CurrentScreen::UrlInput; 
                                                    app.url_input = "https://".to_string();
                                                },
                                                "infra-up" => app.action_run_docker(&["compose", "-f", "deploy/docker-compose.yml", "up", "-d"]),
                                                "infra-down" => app.action_run_docker(&["compose", "-f", "deploy/docker-compose.yml", "down"]),
                                                "view-graph" => app.action_run_browser("http://localhost:7474"),
                                                "check-storage" => app.action_run_browser("http://localhost:9001"),
                                                "check-queue" => app.action_run_browser("http://localhost:15672"),
                                                "run-tests" => {
                                                     app.logs.lock().unwrap().push("[CMD] Running Unit Tests...".to_string());
                                                     let _ = std::process::Command::new("cmd").args(&["/C", "start", "cmd", "/k", "go test ./..."]).current_dir("..").spawn();
                                                },
                                                _ => {}
                                            }
                                        }
                                        _ => {}
                                    }
                                },
                                CurrentScreen::Monitor => {
                                     if key.code == KeyCode::Char('q') { app.should_quit = true; return; }
                                     match key.code {
                                        KeyCode::Tab | KeyCode::Esc => app.screen = CurrentScreen::Table,
                                        _ => {}
                                     }
                                },
                                CurrentScreen::UrlInput => {
                                    match key.code {
                                        KeyCode::Enter => {
                                            app.screen = CurrentScreen::Table; 
                                            if !app.scan_active {
                                                app.scan_active = true;
                                                app.scan_status = "SCANNING".to_string();
                                                app.scan_progress = 0.0;
                                                let target = app.url_input.clone();
                                                let logs_ref = app.logs.clone();
                                                tokio::spawn(run_scanner(logs_ref, target));
                                            }
                                        },
                                        KeyCode::Esc => {
                                            app.screen = CurrentScreen::Table;
                                        },
                                        KeyCode::Char(c) => {
                                            app.url_input.push(c);
                                        },
                                        KeyCode::Backspace => {
                                            app.url_input.pop();
                                        },
                                        _ => {}
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    });

    // Render Loop
    loop {
        {
            let mut app_lock = app.lock().unwrap();
            if app_lock.should_quit {
                break;
            }
            terminal.draw(|f| ui(f, &mut app_lock))?;
        }
        tokio::time::sleep(Duration::from_millis(33)).await;
    }

    disable_raw_mode()?;
    execute!(terminal.backend_mut(), LeaveAlternateScreen)?;
    terminal.show_cursor()?;
    
    Ok(())
}

async fn run_scanner(logs: Arc<Mutex<Vec<String>>>, url: String) {
    logs.lock().unwrap().push(format!("[SYSTEM] Spawning Scanner for {}...", url));
    
    let mut child = Command::new("go")
        .current_dir("..") // Fix CWD
        .args(&["run", "cmd/scanner/main.go", "-url", url.as_str()])
        .stdout(Stdio::piped())
        .spawn();

    match child {
        Ok(mut child) => {
            let stdout = child.stdout.take().unwrap();
            let mut reader = BufReader::new(stdout).lines();
            
            while let Ok(Some(line)) = reader.next_line().await {
                 if let Ok(event) = serde_json::from_str::<BackendEvent>(&line) {
                     // Smart Parsing
                     if let Some(msg) = event.message {
                         let kind = event.kind.unwrap_or("INFO".to_string());
                         logs.lock().unwrap().push(format!("[{}] {}", kind.to_uppercase(), msg));
                     } else if let Some(msg) = event.msg {
                         // Slog handler
                         let lvl = event.level.unwrap_or("LOG".to_string());
                         logs.lock().unwrap().push(format!("[{}] {}", lvl, msg));
                     }
                 } else {
                     logs.lock().unwrap().push(format!("[RAW] {}", line));
                 }
            }
            logs.lock().unwrap().push("[SYSTEM] Scan Process Finished.".to_string());
        }
        Err(e) => {
             logs.lock().unwrap().push(format!("[ERROR] Failed to start scanner: {}", e));
        }
    }
}

// --- UI Rendering ---

fn ui(f: &mut Frame, app: &mut App) {
    // If Login Screen, Render Login Box ONLY
    if app.screen == CurrentScreen::Login {
        render_login(f, app);
        return;
    }

    let chunks = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Length(3), 
            Constraint::Min(0),    
            Constraint::Length(12) 
        ])
        .split(f.area());

    // 1. Header
    let uptime = app.start_time.elapsed().as_secs();
    let header_text = format!(" Context: Admin | PhishVault v7.2 (Secure) | Uptime: {}s | Status: {}", 
         uptime, app.scan_status);
    
    let header = Paragraph::new(header_text)
        .style(Style::default().bg(COLOR_ACCENT).fg(Color::Black).add_modifier(Modifier::BOLD))
        .block(Block::default());
    f.render_widget(header, chunks[0]);

    // 2. Content
    match app.screen {
        CurrentScreen::Table | CurrentScreen::UrlInput => render_table(f, app, chunks[1]),
        CurrentScreen::Monitor => render_charts(f, app, chunks[1]),
        _ => {}
    }

    // 3. Logs
    let logs = app.logs.lock().unwrap();
    let log_text: String = logs.iter().rev().take(10).rev().cloned().collect::<Vec<String>>().join("\n");
    let log_block = Block::default()
        .title(" System Logs ")
        .borders(Borders::ALL)
        .border_type(BorderType::Rounded)
        .border_style(Style::default().fg(COLOR_DIM));
    f.render_widget(Paragraph::new(log_text).style(Style::default().fg(COLOR_ACCENT)).block(log_block), chunks[2]);
    
    // 4. Input Popup (Layered)
    if app.screen == CurrentScreen::UrlInput {
        let block = Block::default().title(" Enter Target URL ").borders(Borders::ALL).border_type(BorderType::Rounded).style(Style::default().fg(COLOR_PRIMARY));
        let area = centered_rect(60, 20, f.area());
        f.render_widget(Clear, area); 
        f.render_widget(Paragraph::new(app.url_input.as_str()).block(block).style(Style::default().fg(COLOR_TEXT)), area);
    }
}

fn render_login(f: &mut Frame, app: &mut App) {
    // Vertically centered, fixed height for crisp look
    let area = centered_rect_fixed(40, 14, f.area());
    
    let block = Block::default()
        .borders(Borders::ALL)
        .border_type(BorderType::Thick)
        .border_style(Style::default().fg(COLOR_WARN));
    
    f.render_widget(Clear, area);
    f.render_widget(block, area);

    let chunks = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Length(1), // Top Padding
            Constraint::Length(2), // Title
            Constraint::Length(3), // User
            Constraint::Length(3), // Pass
            Constraint::Length(2), // Error
            Constraint::Min(0),
        ])
        .margin(2)
        .split(area);

    // Title
    f.render_widget(Paragraph::new("PHISHVAULT COMMAND").alignment(Alignment::Center).style(Style::default().add_modifier(Modifier::BOLD).fg(COLOR_ACCENT)), chunks[1]);

    // Username
    let user_style = if app.login_field == LoginField::Username { Style::default().fg(COLOR_ACCENT) } else { Style::default().fg(COLOR_DIM) };
    let user_block = Block::default().borders(Borders::ALL).title(" ID ").style(user_style);
    f.render_widget(Paragraph::new(app.username_input.as_str()).block(user_block), chunks[2]);

    // Password
    let pass_style = if app.login_field == LoginField::Password { Style::default().fg(COLOR_ACCENT) } else { Style::default().fg(COLOR_DIM) };
    let pass_block = Block::default().borders(Borders::ALL).title(" KEY ").style(pass_style);
    let masked_pass: String = app.password_input.chars().map(|_| '*').collect();
    f.render_widget(Paragraph::new(masked_pass.as_str()).block(pass_block), chunks[3]);

    // Error
    if !app.login_error.is_empty() {
        f.render_widget(Paragraph::new(app.login_error.as_str()).style(Style::default().fg(COLOR_ERROR).add_modifier(Modifier::BOLD)).alignment(Alignment::Center), chunks[4]);
    }
}

fn centered_rect_fixed(percent_x: u16, height: u16, r: Rect) -> Rect {
    let popup_layout = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Length((r.height.saturating_sub(height)) / 2),
            Constraint::Length(height),
            Constraint::Length((r.height.saturating_sub(height)) / 2),
        ])
        .split(r);

    Layout::default()
        .direction(Direction::Horizontal)
        .constraints([
            Constraint::Percentage((100 - percent_x) / 2),
            Constraint::Percentage(percent_x),
            Constraint::Percentage((100 - percent_x) / 2),
        ])
        .split(popup_layout[1])[1]
}

fn render_table(f: &mut Frame, app: &mut App, area: Rect) {
    let header_cells = ["NAME", "CATEGORY", "STATUS", "COMMAND"]
        .iter()
        .map(|h| Cell::from(*h).style(Style::default().fg(COLOR_ACCENT).add_modifier(Modifier::BOLD)));
    let header = Row::new(header_cells).height(1).bottom_margin(1);

    let rows = app.items.iter().map(|item| {
        let cells = vec![
            Cell::from(item.0.clone()),
            Cell::from(item.1.clone()),
            Cell::from(item.2.clone()),
            Cell::from(item.3.clone()),
        ];
        Row::new(cells).height(1)
    });

    let t = Table::new(
        rows,
        [
            Constraint::Percentage(20),
            Constraint::Percentage(20),
            Constraint::Percentage(15),
            Constraint::Percentage(45),
        ]
    )
    .header(header)
    .block(Block::default().borders(Borders::NONE))
    .row_highlight_style(Style::default().bg(COLOR_DIM).add_modifier(Modifier::BOLD)) 
    .highlight_symbol(">> ");
    
    f.render_stateful_widget(t, area, &mut app.table_state);
}

fn render_charts(f: &mut Frame, app: &mut App, area: Rect) {
    let chunks = Layout::default()
        .direction(Direction::Horizontal)
        .constraints([Constraint::Percentage(50), Constraint::Percentage(50)])
        .split(area);

    let datasets = vec![
        Dataset::default()
            .name("CPU Load")
            .marker(symbols::Marker::Braille)
            .style(Style::default().fg(COLOR_ACCENT))
            .graph_type(GraphType::Line)
            .data(&app.cpu_data),
    ];
    let chart = Chart::new(datasets)
        .block(Block::default().title(" Monitor Agent ").borders(Borders::ALL).border_type(BorderType::Rounded))
        .x_axis(Axis::default().title("Time").bounds(app.window).labels(vec![Span::from("-50s"), Span::from("Now")]))
        .y_axis(Axis::default().title("%").bounds([0.0, 100.0]).labels(vec![Span::from("0"), Span::from("100")]));
    f.render_widget(chart, chunks[0]);

    let datasets_mem = vec![
        Dataset::default()
            .name("Memory")
            .marker(symbols::Marker::Block)
            .style(Style::default().fg(COLOR_PRIMARY))
            .graph_type(GraphType::Line)
            .data(&app.mem_data),
    ];
    let chart_mem = Chart::new(datasets_mem)
        .block(Block::default().title(" Orchestrator Load ").borders(Borders::ALL).border_type(BorderType::Rounded))
         .x_axis(Axis::default().bounds(app.window))
         .y_axis(Axis::default().bounds([0.0, 100.0]));
    f.render_widget(chart_mem, chunks[1]);
}

fn centered_rect(percent_x: u16, percent_y: u16, r: Rect) -> Rect {
    let popup_layout = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Percentage((100 - percent_y) / 2),
            Constraint::Percentage(percent_y),
            Constraint::Percentage((100 - percent_y) / 2),
        ])
        .split(r);

    Layout::default()
        .direction(Direction::Horizontal)
        .constraints([
            Constraint::Percentage((100 - percent_x) / 2),
            Constraint::Percentage(percent_x),
            Constraint::Percentage((100 - percent_x) / 2),
        ])
        .split(popup_layout[1])[1]
}
