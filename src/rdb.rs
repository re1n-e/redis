use crate::resp::Value;
use std::collections::HashMap;
use std::time::{Duration, Instant};

#[derive(Debug)]
pub enum RdbError {
    InvalidExpiry(String),
}

struct Set {
    value: Value,
    expiry: Option<Instant>,
}

pub struct Rdb {
    values: HashMap<Value, Set>,
}

impl Rdb {
    pub fn new() -> Self {
        Self {
            values: HashMap::new(),
        }
    }

    pub fn set(
        &mut self,
        key: Value,
        value: Value,
        expiry_type: Option<&Value>,
        expiry_value: Option<&Value>,
    ) -> Result<(), RdbError> {
        let expiry_instant = if let (Some(exp_type), Some(exp_val)) = (expiry_type, expiry_value) {
            let exp_type_str = match exp_type {
                Value::BulkString(s) | Value::SimpleString(s) => s.to_lowercase(),
                _ => {
                    return Err(RdbError::InvalidExpiry(
                        "Expiry type must be a string".to_string(),
                    ))
                }
            };

            let exp_val_str = match exp_val {
                Value::BulkString(s) | Value::SimpleString(s) => s,
                _ => {
                    return Err(RdbError::InvalidExpiry(
                        "Expiry value must be a string".to_string(),
                    ))
                }
            };

            match exp_val_str.parse::<u64>() {
                Ok(duration) => match exp_type_str.as_str() {
                    "ex" => Some(Instant::now() + Duration::from_secs(duration)),
                    "px" => Some(Instant::now() + Duration::from_millis(duration)),
                    _ => {
                        return Err(RdbError::InvalidExpiry(format!(
                            "Unknown expiry type: {}",
                            exp_type_str
                        )))
                    }
                },
                Err(_) => {
                    return Err(RdbError::InvalidExpiry(format!(
                        "Invalid expiry value: {}",
                        exp_val_str
                    )))
                }
            }
        } else {
            None
        };

        println!("expiry: {:?}", expiry_instant);
        self.values.insert(
            key,
            Set {
                value,
                expiry: expiry_instant,
            },
        );
        Ok(())
    }

    pub fn get(&mut self, key: &Value) -> Option<Value> {
        if let Some(entry) = self.values.get_mut(key) {
            if let Some(expiry) = entry.expiry {
                if Instant::now() >= expiry {
                    self.values.remove(key); // remove expired
                    return None;
                }
            }
            return Some(entry.value.clone());
        }
        None
    }
}
