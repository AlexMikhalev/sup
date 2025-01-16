use anyhow::Result;
use clap::Parser;
use std::path::PathBuf;
use tracing::{debug, info};

mod config;
mod executor;

use config::Supfile;
use executor::Executor;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    /// Path to Supfile
    #[arg(short, long, default_value = "Supfile.yml")]
    file: PathBuf,

    /// Network to use
    #[arg(default_value = "dev")]
    network: String,

    /// Command to execute
    #[arg(default_value = "bash")]
    command: String,

    /// Enable debug output
    #[arg(short, long)]
    debug: bool,
}

#[tokio::main]
async fn main() -> Result<()> {
    let args = Args::parse();

    // Initialize logging
    let subscriber = tracing_subscriber::FmtSubscriber::builder()
        .with_max_level(if args.debug { tracing::Level::DEBUG } else { tracing::Level::INFO })
        .with_target(false)
        .with_thread_ids(true)
        .with_file(true)
        .with_line_number(true)
        .with_env_filter("sup_rs=debug")
        .init();

    debug!("Loading Supfile from {}", args.file.display());
    let supfile = Supfile::from_file(&args.file)?;

    let network = supfile.networks.get(&args.network)
        .ok_or_else(|| anyhow::anyhow!("Network {} not found", args.network))?;

    let command = supfile.commands.get(&args.command)
        .ok_or_else(|| anyhow::anyhow!("Command {} not found", args.command))?;

    let mut env = std::env::vars().collect::<std::collections::HashMap<_, _>>();
    if let Some(vars) = &supfile.env {
        env.extend(vars.clone());
    }

    let executor = Executor::new(network.clone(), env);
    executor.execute_command(command).await?;

    Ok(())
} 