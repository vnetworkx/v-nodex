use crate::record::VectorRecord;
use std::collections::BTreeMap;

#[derive(Default, Clone, Debug)]
pub struct MemoryRecordStore {
    records: BTreeMap<String, VectorRecord>,
}

impl MemoryRecordStore {
    pub fn append(&mut self, record: VectorRecord) {
        self.records.insert(record.record_id.clone(), record);
    }

    pub fn get(&self, record_id: &str) -> Option<&VectorRecord> {
        self.records.get(record_id)
    }

    pub fn replay_order(&self) -> Vec<&VectorRecord> {
        self.records.values().collect()
    }
}
