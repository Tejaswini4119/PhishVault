use anyhow::Result;
use crossterm::{
    event::{Event, KeyCode, KeyEventKind},
    execute,
    terminal::{disable_raw_mode, enable_raw_mode, EnterAlternateScreen, LeaveAlternateScreen},
};
use futures::StreamExt;
use ratatui::{
    prelude::*,
    Terminal,
};
use std::io;
use std::sync::{Arc, Mutex};
use std::time::Duration;
use tokio::sync::mpsc; // Import mpsc

mod modules;
use modules::app::{App, CurrentScreen, LoginField};
use modules::ui::ui;
use modules::scanner::run_scanner;
use modules::logger::LogEntry;
use modules::events::AppEvent; // Import AppEvent

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
    
    // Create Internal Event Channel
    let (tx, mut rx) = mpsc::unbounded_channel::<AppEvent>();

    // Event Loop
    let _input_handle = tokio::spawn(async move {
        let mut tick_interval = tokio::time::interval(Duration::from_millis(100));
        let mut reader = crossterm::event::EventStream::new();
        let tx_scanner = tx.clone(); // Generic sender clone for tasks

        loop {
            tokio::select! {
                // 1. Ticks
                _ = tick_interval.tick() => {
                    let mut app = app_clone.lock().unwrap();
                    app.on_tick();
                }
                // 2. Internal Events (Async Tasks)
                Some(internal_event) = rx.recv() => {
                     match internal_event {
                         AppEvent::ScanComplete => {
                             let mut app = app_clone.lock().unwrap();
                             if app.scan_active {
                                 app.scan_active = false;
                                 app.scan_status = "COMPLETE".to_string();
                                 app.scan_progress = 1.0;
                                 // Ensure we stop the wave animation
                             }
                         }
                     }
                }
                // 3. User Input (Terminal)
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
                                                        app.logs.lock().unwrap().push(LogEntry::new("SUCCESS","Admin Authenticated."));
                                                        app.login_error.clear();
                                                    } else {
                                                        app.login_error = "Invalid Credentials!".to_string();
                                                        app.password_input.clear();
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
                                                     app.logs.lock().unwrap().push(LogEntry::new("CMD", "Running Unit Tests..."));
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
                                                let tx_ref = tx_scanner.clone(); // Pass Sender to Scanner task
                                                tokio::spawn(run_scanner(logs_ref, target, tx_ref));
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
