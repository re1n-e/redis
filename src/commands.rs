use crate::resp::Value;

pub enum CommandType {
    Ping,
    Echo,
    Set,
    Get,
    Rpush,
    LPush,
    Lrange,
    LLen,
    LPop,
    BLpop,
    Unknown,
}

impl From<&str> for CommandType {
    fn from(type_str: &str) -> Self {
        match type_str {
            "PING" => CommandType::Ping,
            "ECHO" => CommandType::Echo,
            "SET" => CommandType::Set,
            "GET" => CommandType::Get,
            "RPUSH" => CommandType::Rpush,
            "LPUSH" => CommandType::LPush,
            "LRANGE" => CommandType::Lrange,
            "LLEN" => CommandType::LLen,
            "LPOP" => CommandType::LPop,
            "BLPOP" => CommandType::BLpop,
            _ => CommandType::Unknown,
        }
    }
}

pub struct Command {
    pub command_type: CommandType,
    pub args: Vec<Value>,
}

impl Command {
    pub fn new(command: &str, args: Vec<Value>) -> Self {
        Command {
            command_type: CommandType::from(command),
            args,
        }
    }
}
