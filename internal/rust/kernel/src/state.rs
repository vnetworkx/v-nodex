use crate::{record::VectorRecord, types::{AuthRatio, Vector, VectorType}};
use std::collections::BTreeMap;

#[derive(Clone, Debug, PartialEq)]
pub struct WalletState {
    pub wallet_id: String,
    pub public_key: String,
    pub vector: Vector,
    pub vector_type: VectorType,
    pub certification: AuthRatio,
    pub latest_record_id: Option<String>,
    pub metadata: BTreeMap<String, String>,
}

impl WalletState {
    pub fn new(wallet_id: impl Into<String>, public_key: impl Into<String>, vector: Vector) -> Self {
        Self {
            wallet_id: wallet_id.into(),
            public_key: public_key.into(),
            vector,
            vector_type: VectorType::Free,
            certification: AuthRatio(1.0),
            latest_record_id: None,
            metadata: BTreeMap::new(),
        }
    }
}

#[derive(Clone, Debug, PartialEq)]
pub struct KernelState {
    pub protocol_version: String,
    pub wallets: BTreeMap<String, WalletState>,
    pub records: BTreeMap<String, VectorRecord>,
}

impl Default for KernelState {
    fn default() -> Self {
        Self {
            protocol_version: "1.0".to_string(),
            wallets: BTreeMap::new(),
            records: BTreeMap::new(),
        }
    }
}

impl KernelState {
    pub fn upsert_wallet(&mut self, wallet: WalletState) {
        self.wallets.insert(wallet.wallet_id.clone(), wallet);
    }

    pub fn wallet(&self, id: &str) -> Option<&WalletState> {
        self.wallets.get(id)
    }
}
