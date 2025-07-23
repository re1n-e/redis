use crate::resp::Value;

pub enum CommandType {
    Ping,
    Echo,
    Set,
    Get,
    Unknown,
}

impl From<&str> for CommandType {
    fn from(type_str: &str) -> Self {
        match type_str {
            "PING" => CommandType::Ping,
            "ECHO" => CommandType::Echo,
            "SET" => CommandType::Set,
            "GET" => CommandType::Get,
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
