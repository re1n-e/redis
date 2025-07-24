use crate::commands::{Command, CommandType};
use crate::rdb::Rdb;
use crate::resp::{RespHandler, Value};

use std::collections::HashMap;

pub struct Redis {
    rdb: Rdb,
    lists: HashMap<String, Vec<Value>>,
}

impl Redis {
    pub fn new() -> Self {
        Redis {
            rdb: Rdb::new(),
            lists: HashMap::new(),
        }
    }

    pub async fn execute_command(&mut self, cmd: &Command, handler: &mut RespHandler) {
        let response = match cmd.command_type {
            CommandType::Ping => Value::SimpleString("PONG".into()),

            CommandType::Echo => cmd.args.get(0).cloned().unwrap_or_else(|| {
                Value::SimpleString("ERR wrong number of arguments for 'echo' command".into())
            }),

            CommandType::Get => match self.rdb.get(&cmd.args[0]) {
                Some(val) => Value::BulkString(format!("{val}")),
                None => Value::NullBulkString,
            },

            CommandType::Set => {
                let expiry = cmd.args.get(3);
                self.rdb
                    .set(
                        cmd.args[0].clone(),
                        cmd.args[1].clone(),
                        cmd.args.get(2),
                        expiry,
                    )
                    .unwrap();
                Value::SimpleString("OK".to_string())
            }

            CommandType::Rpush => {
                let key = match &cmd.args[0] {
                    Value::BulkString(s) | Value::SimpleString(s) => s.clone(),
                    _ => panic!("Invalid key for RPUSH"),
                };

                let list = self.lists.entry(key).or_insert_with(Vec::new);
                for arg in &cmd.args[1..] {
                    list.push(arg.clone());
                }
                Value::IntegerString(list.len())
            }

            CommandType::LPush => {
                let key = match &cmd.args[0] {
                    Value::BulkString(s) | Value::SimpleString(s) => s.clone(),
                    _ => panic!("Invalid key for LPUSH"),
                };

                let list = self.lists.entry(key).or_insert_with(Vec::new);
                for arg in &cmd.args[1..] {
                    list.insert(0, arg.clone()); // O(n) shift
                }
                Value::IntegerString(list.len())
            }

            CommandType::Lrange => {
                let key = match &cmd.args[0] {
                    Value::BulkString(s) | Value::SimpleString(s) => s.clone(),
                    _ => panic!("Invalid key for LRANGE"),
                };

                let result = self.lists.get(&key).map_or(vec![], |list| {
                    let len = list.len() as isize;

                    let mut start = match cmd.args.get(1) {
                        Some(Value::BulkString(s)) => s.parse::<isize>().unwrap_or(0),
                        _ => 0,
                    };

                    let mut end = match cmd.args.get(2) {
                        Some(Value::BulkString(s)) => s.parse::<isize>().unwrap_or(0),
                        _ => 0,
                    };

                    if start < 0 {
                        start += len;
                    }
                    if end < 0 {
                        end += len;
                    }

                    start = start.max(0);
                    end = end.min(len - 1);

                    if start > end || start >= len {
                        vec![]
                    } else {
                        list[start as usize..=end as usize].to_vec()
                    }
                });

                Value::Array(result)
            }

            CommandType::LLen => {
                let key = match &cmd.args[0] {
                    Value::BulkString(s) | Value::SimpleString(s) => s.clone(),
                    _ => panic!("Invalid key for LLEN"),
                };

                let len = self.lists.get(&key).map_or(0, |list| list.len());
                Value::IntegerString(len)
            }
            CommandType::LPop => {
                let key = match &cmd.args[0] {
                    Value::BulkString(s) | Value::SimpleString(s) => s.clone(),
                    _ => panic!("Invalid key for LPOP"),
                };

                let list = self.lists.entry(key).or_insert_with(Vec::new);

                if list.is_empty() {
                    Value::NullBulkString
                } else if cmd.args.len() == 1 {
                    let value = list.remove(0);
                    value
                } else {
                    let count = match cmd.args.get(1) {
                        Some(Value::BulkString(s)) => s.parse::<usize>().unwrap_or(1),
                        _ => 1,
                    };

                    let actual_count = count.min(list.len());
                    let mut result = Vec::with_capacity(actual_count);

                    for _ in 0..actual_count {
                        result.push(list.remove(0));
                    }

                    Value::Array(result)
                }
            }
            CommandType::BLpop => {}
            CommandType::Unknown => Value::SimpleString("ERR unknown command".into()),
        };

        println!("Sending value {:?}", response);
        handler.write_value(response).await.unwrap();
    }
}
