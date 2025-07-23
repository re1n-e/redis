use crate::commands::Command;
use crate::redis::Redis;
use crate::resp::{RespHandler, Value};
use anyhow::Result;
use std::sync::Arc;
use tokio::net::TcpStream;
use tokio::sync::Mutex;

pub async fn handle_conn(stream: TcpStream, redis: Arc<Mutex<Redis>>) {
    let mut handler = RespHandler::new(stream);
    println!("Starting read loop");

    loop {
        let value = match handler.read_value().await {
            Ok(Some(v)) => v,
            Ok(None) => break, // client closed connection
            Err(e) => {
                eprintln!("Read error: {e}");
                break;
            }
        };

        println!("Got value {:?}", value);

        let (cmd, args) = extract_command(value).unwrap();
        let command = Command::new(&cmd, args);
        let mut redis = redis.lock().await;
        redis.execute_command(&command, &mut handler).await;
    }
}

fn extract_command(value: Value) -> Result<(String, Vec<Value>)> {
    match value {
        Value::Array(a) => Ok((
            unpack_bulk_str(a.first().unwrap().clone())?,
            a.into_iter().skip(1).collect(),
        )),
        _ => Err(anyhow::anyhow!("Unexpected command format")),
    }
}

fn unpack_bulk_str(value: Value) -> Result<String> {
    match value {
        Value::BulkString(s) => Ok(s),
        _ => Err(anyhow::anyhow!("Expected command to be a bulk string")),
    }
}
