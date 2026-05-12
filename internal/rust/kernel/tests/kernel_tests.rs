use vnodex_kernel::{apply_operation, KernelState, Operation, Vector, VectorType};

#[test]
fn normalize_zero_is_rejected() {
    let mut state = KernelState::default();
    let result = apply_operation(&mut state, "wallet-1", Operation::Normalize, Vector::zero(3), None, None, Some(VectorType::Free));
    assert!(result.is_err());
}

#[test]
fn add_operation_updates_state() {
    let mut state = KernelState::default();
    let result = apply_operation(&mut state, "wallet-1", Operation::Create, Vector::new(vec![1.0, 2.0]), None, None, Some(VectorType::Free));
    assert!(result.is_ok());
    let result = apply_operation(&mut state, "wallet-1", Operation::Add, Vector::new(vec![2.0, 3.0]), None, None, None);
    assert!(result.is_ok());
    let wallet = state.wallet("wallet-1").unwrap();
    assert_eq!(wallet.vector.components, vec![3.0, 5.0]);
}
