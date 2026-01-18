use std::sync::{Arc, Mutex};
use std::process::Stdio;
use tokio::process::Command;
use tokio::io::{AsyncBufReadExt, BufReader};
use tokio::sync::mpsc::UnboundedSender;
use crate::modules::events::{BackendEvent, AppEvent};
use crate::modules::logger::LogEntry;

pub async fn run_scanner(logs: Arc<Mutex<Vec<LogEntry>>>, url: String, tx: UnboundedSender<AppEvent>) {
    logs.lock().unwrap().push(LogEntry::new("SYSTEM", &format!("Spawning Scanner for {}...", url)));
    
    let child = Command::new("go")
        .current_dir("..") // Fix CWD to project root
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
                         logs.lock().unwrap().push(LogEntry::new(&kind.to_uppercase(), &msg));
                     } else if let Some(msg) = event.msg {
                         // Slog handler
                         let lvl = event.level.unwrap_or("LOG".to_string());
                         logs.lock().unwrap().push(LogEntry::new(&lvl.to_uppercase(), &msg));
                     }
                 } else {
                     logs.lock().unwrap().push(LogEntry::new("RAW", &line));
                 }
            }
            logs.lock().unwrap().push(LogEntry::new("SYSTEM", "Scan Process Finished."));
            let _ = tx.send(AppEvent::ScanComplete);
        }
        Err(e) => {
             logs.lock().unwrap().push(LogEntry::new("ERROR", &format!("Failed to start scanner: {}", e)));
             let _ = tx.send(AppEvent::ScanComplete);
        }
    }
}
