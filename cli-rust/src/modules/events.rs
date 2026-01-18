use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct BackendEvent {
    #[serde(rename = "type")]
    pub kind: Option<String>,
    pub message: Option<String>,
    pub level: Option<String>,
    pub msg: Option<String>,
}

#[derive(Debug, Clone)]
pub enum AppEvent {
    ScanComplete,
}
