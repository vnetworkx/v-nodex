use crate::{operations::Operation, types::Vector};
use std::collections::BTreeMap;

#[derive(Clone, Debug, PartialEq)]
pub struct VectorRecord {
    pub record_id: String,
    pub prev_hash: Option<String>,
    pub v_before: Vector,
    pub v_after: Vector,
    pub operation: Operation,
    pub params: BTreeMap<String, String>,
    pub auth_ratio: f64,
    pub certified: bool,
    pub actor_pk: String,
    pub proof: String,
    pub timestamp: u64,
    pub signature: String,
}

impl VectorRecord {
    pub fn new(record_id: impl Into<String>, operation: Operation, before: Vector, after: Vector) -> Self {
        Self {
            record_id: record_id.into(),
            prev_hash: None,
            v_before: before,
            v_after: after,
            operation,
            params: BTreeMap::new(),
            auth_ratio: 1.0,
            certified: true,
            actor_pk: String::new(),
            proof: String::new(),
            timestamp: 0,
            signature: String::new(),
        }
    }
}
