use ratatui::{
    prelude::*,
    widgets::{Block, BorderType, Borders, Cell, Clear, Paragraph, Row, Table, Dataset, Chart, Axis, GraphType},
    symbols,
};
use crate::modules::app::{App, CurrentScreen, LoginField};
use crate::modules::theme::{ACCENT, PRIMARY, DIM, TEXT, WARN, ERROR, SUCCESS};

pub fn ui(f: &mut Frame, app: &mut App) {
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
        .style(Style::default().bg(ACCENT).fg(Color::Black).add_modifier(Modifier::BOLD))
        .block(Block::default());
    f.render_widget(header, chunks[0]);

    // 2. Content
    match app.screen {
        CurrentScreen::Table | CurrentScreen::UrlInput => render_table(f, app, chunks[1]),
        CurrentScreen::Monitor => render_charts(f, app, chunks[1]),
        _ => {}
    }

    // 3. Logs (Structured Rendering)
    let logs = app.logs.lock().unwrap();
    let visible_logs: Vec<Line> = logs.iter().rev().take(10).rev().map(|entry| {
        let style = match entry.level.as_str() {
            "ERROR" => Style::default().fg(ERROR),
            "WARN" => Style::default().fg(WARN),
            "SUCCESS" => Style::default().fg(SUCCESS),
            "CMD" => Style::default().fg(PRIMARY),
            "SYSTEM" => Style::default().fg(DIM),
            "DOCKER" => Style::default().fg(Color::Cyan),
            _ => Style::default().fg(TEXT),
        };
        
        Line::from(vec![
            Span::styled(format!("[{}] ", entry.level), style.add_modifier(Modifier::BOLD)),
            Span::raw(entry.message.clone()),
        ])
    }).collect();

    let log_block = Block::default()
        .title(" System Logs ")
        .borders(Borders::ALL)
        .border_type(BorderType::Rounded)
        .border_style(Style::default().fg(DIM));
    
    f.render_widget(Paragraph::new(visible_logs).block(log_block), chunks[2]);
    
    // 4. Input Popup (Layered)
    if app.screen == CurrentScreen::UrlInput {
        let block = Block::default().title(" Enter Target URL ").borders(Borders::ALL).border_type(BorderType::Rounded).style(Style::default().fg(PRIMARY));
        let area = centered_rect(60, 20, f.area());
        f.render_widget(Clear, area); 
        f.render_widget(Paragraph::new(app.url_input.as_str()).block(block).style(Style::default().fg(TEXT)), area);
    }
}

fn render_login(f: &mut Frame, app: &mut App) {
    // Vertically centered, fixed height for crisp look
    let area = centered_rect_fixed(40, 14, f.area());
    
    let block = Block::default()
        .borders(Borders::ALL)
        .border_type(BorderType::Thick)
        .border_style(Style::default().fg(WARN));
    
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
    f.render_widget(Paragraph::new("PHISHVAULT COMMAND").alignment(Alignment::Center).style(Style::default().add_modifier(Modifier::BOLD).fg(ACCENT)), chunks[1]);

    // Username
    let user_style = if app.login_field == LoginField::Username { Style::default().fg(ACCENT) } else { Style::default().fg(DIM) };
    let user_block = Block::default().borders(Borders::ALL).title(" ID ").style(user_style);
    f.render_widget(Paragraph::new(app.username_input.as_str()).block(user_block), chunks[2]);

    // Password
    let pass_style = if app.login_field == LoginField::Password { Style::default().fg(ACCENT) } else { Style::default().fg(DIM) };
    let pass_block = Block::default().borders(Borders::ALL).title(" KEY ").style(pass_style);
    let masked_pass: String = app.password_input.chars().map(|_| '*').collect();
    f.render_widget(Paragraph::new(masked_pass.as_str()).block(pass_block), chunks[3]);

    // Error
    if !app.login_error.is_empty() {
        f.render_widget(Paragraph::new(app.login_error.as_str()).style(Style::default().fg(ERROR).add_modifier(Modifier::BOLD)).alignment(Alignment::Center), chunks[4]);
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

fn render_table(f: &mut Frame, app: &mut App, area: Rect) {
    let header_cells = ["NAME", "CATEGORY", "STATUS", "COMMAND"]
        .iter()
        .map(|h| Cell::from(*h).style(Style::default().fg(ACCENT).add_modifier(Modifier::BOLD)));
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
    .block(Block::default().borders(Borders::ALL).border_type(BorderType::Rounded).border_style(Style::default().fg(DIM)))
    .row_highlight_style(Style::default().bg(DIM).add_modifier(Modifier::BOLD)) 
    .highlight_symbol("► ");
    
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
            .style(Style::default().fg(ACCENT))
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
            .style(Style::default().fg(PRIMARY))
            .graph_type(GraphType::Line)
            .data(&app.mem_data),
    ];
    let chart_mem = Chart::new(datasets_mem)
        .block(Block::default().title(" Orchestrator Load ").borders(Borders::ALL).border_type(BorderType::Rounded))
         .x_axis(Axis::default().bounds(app.window))
         .y_axis(Axis::default().bounds([0.0, 100.0]));
    f.render_widget(chart_mem, chunks[1]);
}
