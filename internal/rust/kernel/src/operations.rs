use crate::{errors::KernelError, state::{KernelState, WalletState}, types::{Vector, VectorType}, validation::{validate_normalization, validate_type}};

#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Operation {
    Create,
    Certify,
    Transfer,
    Drain,
    Project,
    Reconstruct,
    Query,
    Record,
    Add,
    Subtract,
    Scale,
    Normalize,
    Rotate,
    Constrain,
    Compose,
    Nullify,
}

#[derive(Clone, Debug, PartialEq)]
pub struct OperationResult {
    pub before: Vector,
    pub after: Vector,
    pub certified: bool,
    pub message: String,
}

pub fn apply_operation(
    state: &mut KernelState,
    wallet_id: &str,
    operation: Operation,
    input: Vector,
    secondary: Option<Vector>,
    scalar: Option<f64>,
    vector_type: Option<VectorType>,
) -> Result<OperationResult, KernelError> {
    let before = state
        .wallet(wallet_id)
        .map(|w| w.vector.clone())
        .unwrap_or_else(|| Vector::zero(input.dimension()));

    let after = match operation {
        Operation::Create => {
            validate_type(vector_type.as_ref().unwrap_or(&VectorType::Free), &input)?;
            input.clone()
        }
        Operation::Certify => before.clone(),
        Operation::Transfer | Operation::Subtract => before.sub(&input),
        Operation::Drain => before.scale(1.0 - scalar.unwrap_or(0.0)),
        Operation::Project => {
            let axis = secondary.ok_or(KernelError::InvalidOperation("project requires secondary vector"))?;
            before.project_onto(&axis)?
        }
        Operation::Reconstruct | Operation::Query | Operation::Record => before.clone(),
        Operation::Add | Operation::Compose => before.add(&input),
        Operation::Scale => before.scale(scalar.unwrap_or(1.0)),
        Operation::Normalize => {
            validate_normalization(&before)?;
            before.normalize()?
        }
        Operation::Rotate => {
            let theta = scalar.unwrap_or(0.0);
            before.rotate_2d(theta)?
        }
        Operation::Constrain => {
            let bounds = secondary.ok_or(KernelError::InvalidOperation("constrain requires bounds vector"))?;
            before.constrain(&Vector::zero(bounds.dimension()), &bounds)
        }
        Operation::Nullify => Vector::zero(before.dimension()),
    };

    let wallet = state
        .wallets
        .entry(wallet_id.to_string())
        .or_insert_with(|| WalletState::new(wallet_id.to_string(), String::new(), before.clone()));
    wallet.vector = after.clone();
    wallet.vector_type = vector_type.unwrap_or_else(|| wallet.vector_type.clone());

    Ok(OperationResult {
        before,
        after,
        certified: true,
        message: format!("applied {:?}", operation),
    })
}
