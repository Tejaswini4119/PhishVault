use ratatui::widgets::TableState;
use std::sync::{Arc, Mutex};
use std::time::Instant;
use tokio::io::{AsyncBufReadExt, BufReader};
use tokio::process::Command;
use std::process::Stdio;
use crate::modules::logger::LogEntry;

#[derive(PartialEq, Clone, Copy, Debug)]
pub enum CurrentScreen {
    Login,
    Table,
    Monitor,
    UrlInput, 
}

#[derive(PartialEq, Clone, Copy, Debug)]
pub enum LoginField {
    Username,
    Password,
}

pub struct App {
    pub screen: CurrentScreen,
    pub logs: Arc<Mutex<Vec<LogEntry>>>, // Changed to LogEntry
    
    // Login State
    pub login_field: LoginField,
    pub username_input: String,
    pub password_input: String,
    pub login_error: String,
    
    // Scanner
    pub scan_active: bool,
    pub scan_progress: f64,
    pub scan_status: String,
    pub url_input: String, 
    
    // Monitor Data
    pub cpu_data: Vec<(f64, f64)>,
    pub mem_data: Vec<(f64, f64)>,
    pub window: [f64; 2],

    // Table
    pub table_state: TableState,
    pub items: Vec<(String, String, String, String)>,
    pub start_time: Instant,
    pub should_quit: bool,
}

impl App {
    pub fn new() -> App {
        let mut table_state = TableState::default();
        table_state.select(Some(0));

        App {
            screen: CurrentScreen::Login,
            logs: Arc::new(Mutex::new(vec![LogEntry::new("INFO", "Authentication Required.")])),
            
            login_field: LoginField::Username,
            username_input: String::new(),
            password_input: String::new(),
            login_error: String::new(),

            scan_active: false,
            scan_progress: 0.0,
            scan_status: "LOCKED".to_string(), 
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
    
    pub fn on_tick(&mut self) {
        if self.screen == CurrentScreen::Login { return; } 
        
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

    pub fn next_row(&mut self) {
        let i = match self.table_state.selected() {
            Some(i) => if i >= self.items.len() - 1 { 0 } else { i + 1 },
            None => 0,
        };
        self.table_state.select(Some(i));
    }

    pub fn prev_row(&mut self) {
        let i = match self.table_state.selected() {
            Some(i) => if i == 0 { self.items.len() - 1 } else { i - 1 },
            None => 0,
        };
        self.table_state.select(Some(i));
    }
    
    pub fn action_run_browser(&mut self, url: &str) {
         self.logs.lock().unwrap().push(LogEntry::new("CMD", &format!("Opening Browser: {}", url)));
         let _ = std::process::Command::new("cmd").args(&["/C", "start", url]).spawn();
    }
    
    pub fn action_run_docker(&mut self, args: &[&str]) {
         let args_vec: Vec<String> = args.iter().map(|s| s.to_string()).collect();
         let logs_ref = self.logs.clone();
         
         self.logs.lock().unwrap().push(LogEntry::new("CMD", &format!("Running Docker: {:?}", args)));
         
         tokio::spawn(async move {
             run_docker_async(logs_ref, args_vec).await;
         });
    }
}

pub async fn run_docker_async(logs: Arc<Mutex<Vec<LogEntry>>>, args: Vec<String>) {
    // Robust Path Finding
    let current_dir = std::env::current_dir().unwrap_or_else(|_| std::path::PathBuf::from("."));
    let mut project_root = current_dir.clone();
    let mut found = false;
    
    for _ in 0..5 {
        if project_root.join("deploy").join("docker-compose.yml").exists() {
            found = true;
            break;
        }
        if !project_root.pop() { break; }
    }
    
    if !found {
        logs.lock().unwrap().push(LogEntry::new("WARN", "Could not auto-detect project root. using default '..' path."));
        project_root = std::path::PathBuf::from(".."); 
    } else {
         logs.lock().unwrap().push(LogEntry::new("DEBUG", &format!("Project Root Found: {:?}", project_root)));
    }

    let child = Command::new("docker")
        .args(&args)
        .current_dir(project_root)
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn();

    match child {
        Ok(mut child) => {
             if let Some(stdout) = child.stdout.take() {
                 let logs = logs.clone();
                 tokio::spawn(async move {
                     let mut reader = BufReader::new(stdout).lines();
                     while let Ok(Some(line)) = reader.next_line().await {
                         logs.lock().unwrap().push(LogEntry::new("DOCKER", &line));
                     }
                 });
             }
             if let Some(stderr) = child.stderr.take() {
                 let logs = logs.clone();
                 tokio::spawn(async move {
                     let mut reader = BufReader::new(stderr).lines();
                     while let Ok(Some(line)) = reader.next_line().await {
                         logs.lock().unwrap().push(LogEntry::new("DOCKER", &line));
                     }
                 });
             }
             
             let _ = child.wait().await;
             logs.lock().unwrap().push(LogEntry::new("DOCKER", &format!("Command Finished: {:?}", args)));
        }
        Err(e) => {
             logs.lock().unwrap().push(LogEntry::new("ERROR", &format!("Docker failed: {}", e)));
        }
    }
}
