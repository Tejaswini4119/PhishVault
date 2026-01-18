use std::time::Instant;

#[derive(Clone, Debug)]
pub struct LogEntry {
    #[allow(dead_code)]
    pub timestamp: Instant,
    pub level: String,
    pub message: String,
}

impl LogEntry {
    pub fn new(level: &str, message: &str) -> Self {
        LogEntry {
            timestamp: Instant::now(),
            level: level.to_string(),
            message: message.to_string(),
        }
    }
}
