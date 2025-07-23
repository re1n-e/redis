use std::collections::HashMap;

use crate::resp::Value;

pub struct Rdb {
    values: HashMap<Value, Value>,
}

impl Rdb {
    pub fn new() -> Self {
        Rdb {
            values: HashMap::new(),
        }
    }

    pub fn set(&mut self, key: Value, value: Value) {
        self.values.insert(key, value);
    }

    pub fn get(&self, key: &Value) -> Option<Value> {
        if let Some(val) = self.values.get(key) {
            return Some(val.clone());
        }
        return None;
    }
}
