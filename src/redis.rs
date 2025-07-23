use crate::commands::{Command, CommandType};
use crate::rdb::Rdb;
use crate::resp::{RespHandler, Value};

pub struct Redis {
    rdb: Rdb,
}

impl Redis {
    pub fn new() -> Self {
        Redis { rdb: Rdb::new() }
    }

    pub async fn execute_command(&mut self, cmd: &Command, handler: &mut RespHandler) {
        let response = match cmd.command_type {
            CommandType::Ping => Value::SimpleString("PONG".to_string()),
            CommandType::Echo => cmd.args.first().unwrap().clone(),
            CommandType::Get => match self.rdb.get(&cmd.args[1]) {
                Some(val) => Value::BulkString(format!("{val}")),
                None => Value::SimpleString("$-1\r\n".to_string()),
            },
            CommandType::Set => {
                self.rdb.set(cmd.args[1].clone(), cmd.args[2].clone());
                Value::SimpleString("+OK\r\n".to_string())
            }
            CommandType::Unknown => panic!("Cannot handle command"),
        };
        println!("Sending value {:?}", response);
        handler.write_value(response).await.unwrap();
    }
}
