use std::{
    env,
    error::Error,
    fs, io,
    path::{Path, PathBuf},
};
use tracing::level_filters::LevelFilter;
use tracing_appender::{
    non_blocking::WorkerGuard,
    rolling::{RollingFileAppender, Rotation},
};
use tracing_subscriber::fmt;
use tracing_subscriber::prelude::*;

use crate::config::get_program_dir;

pub fn init(path: impl AsRef<Path>) -> Result<WorkerGuard, Box<dyn Error + Send + Sync>> {
    let appender = RollingFileAppender::builder()
        .rotation(Rotation::DAILY)
        .filename_suffix("log")
        .max_log_files(7)
        .build(path)?;

    let (writer, guard) = tracing_appender::non_blocking(appender);

    let console_layer = fmt::layer().with_writer(std::io::stdout).with_ansi(true);

    let file_layer = fmt::layer().with_writer(writer).with_ansi(false);

    let level = match env::var("LOG_LEVEL").as_deref() {
        Ok("trace") => LevelFilter::TRACE,
        Ok("debug") => LevelFilter::DEBUG,
        Ok("warn") => LevelFilter::WARN,
        Ok("error") => LevelFilter::ERROR,
        Ok("off") => LevelFilter::OFF,
        _ => LevelFilter::INFO,
    };

    tracing_subscriber::registry()
        .with(level)
        .with(console_layer)
        .with(file_layer)
        .try_init()?;

    Ok(guard)
}

pub fn get_log_dir() -> io::Result<PathBuf> {
    let base = get_program_dir();

    let path = base.join("logs");

    fs::create_dir_all(&path)?;

    Ok(path)
}
