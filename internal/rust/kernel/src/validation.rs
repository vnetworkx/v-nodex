use crate::{errors::KernelError, record::VectorRecord, types::{Vector, VectorType}};

pub fn validate_vector(vector: &Vector) -> Result<(), KernelError> {
    if vector.components.is_empty() {
        return Err(KernelError::EmptyVector);
    }
    Ok(())
}

pub fn validate_normalization(vector: &Vector) -> Result<(), KernelError> {
    if vector.is_zero() {
        return Err(KernelError::ZeroNormalization);
    }
    Ok(())
}

pub fn validate_type(vector_type: &VectorType, vector: &Vector) -> Result<(), KernelError> {
    match vector_type {
        VectorType::Zero if !vector.is_zero() => Err(KernelError::InvalidStructure("zero vector type requires zero value")),
        VectorType::Unit if (vector.euclidean_norm() - 1.0).abs() > 1e-9 => Err(KernelError::InvalidStructure("unit vector must have euclidean norm 1")),
        _ => Ok(()),
    }
}

pub fn validate_record(record: &VectorRecord) -> Result<(), KernelError> {
    validate_vector(&record.v_before)?;
    validate_vector(&record.v_after)?;
    Ok(())
}
