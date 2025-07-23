use crate::redis::Redis;
use std::process::exit;
use std::sync::Arc;
use tokio::net::TcpListener;
use tokio::sync::Mutex;
mod commands;
mod handler;
mod rdb;
mod redis;
mod resp;

#[tokio::main]
async fn main() {
    let listener = match TcpListener::bind("127.0.0.1:6379").await {
        Ok(listener) => listener,
        Err(e) => {
            eprintln!("Failed to bind: {e}");
            exit(1);
        }
    };
    let redis = Arc::new(Mutex::new(Redis::new()));
    println!("Listening on 127.0.0.1:6379");

    loop {
        let stream = listener.accept().await;
        let redis = redis.clone();

        match stream {
            Ok((stream, sock_addr)) => {
                println!("Accepted new connection from {sock_addr}");

                tokio::spawn(async move { handler::handle_conn(stream, redis).await });
            }
            Err(e) => println!("error: {e}"),
        }
    }
}
