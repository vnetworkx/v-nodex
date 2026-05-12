use core::fmt;

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum KernelError {
    EmptyVector,
    ZeroNormalization,
    InvalidOperation(&'static str),
    InvalidSignature,
    InvalidStructure(&'static str),
    MissingParent,
    Unauthorized,
    CertificationFailed,
    ReplayMismatch,
}

impl fmt::Display for KernelError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            KernelError::EmptyVector => write!(f, "empty vector"),
            KernelError::ZeroNormalization => write!(f, "cannot normalize zero vector"),
            KernelError::InvalidOperation(msg) => write!(f, "invalid operation: {msg}"),
            KernelError::InvalidSignature => write!(f, "invalid signature"),
            KernelError::InvalidStructure(msg) => write!(f, "invalid structure: {msg}"),
            KernelError::MissingParent => write!(f, "missing parent"),
            KernelError::Unauthorized => write!(f, "unauthorized"),
            KernelError::CertificationFailed => write!(f, "certification failed"),
            KernelError::ReplayMismatch => write!(f, "replay mismatch"),
        }
    }
}

impl std::error::Error for KernelError {}
