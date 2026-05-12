use crate::errors::KernelError;

pub type Component = f64;

#[derive(Clone, Debug, PartialEq)]
pub struct Vector {
    pub components: Vec<Component>,
}

impl Vector {
    pub fn new(components: Vec<Component>) -> Self {
        Self { components }
    }

    pub fn zero(dimensions: usize) -> Self {
        Self { components: vec![0.0; dimensions] }
    }

    pub fn dimension(&self) -> usize {
        self.components.len()
    }

    pub fn is_zero(&self) -> bool {
        self.components.iter().all(|v| *v == 0.0)
    }

    pub fn magnitude(&self) -> f64 {
        self.components.iter().copied().sum()
    }

    pub fn euclidean_norm(&self) -> f64 {
        self.components.iter().map(|v| v * v).sum::<f64>().sqrt()
    }

    pub fn direction(&self) -> Self {
        let total = self.magnitude();
        if total == 0.0 {
            return Self::zero(self.dimension());
        }
        Self {
            components: self.components.iter().map(|v| v / total).collect(),
        }
    }

    pub fn scale(&self, k: f64) -> Self {
        Self {
            components: self.components.iter().map(|v| v * k).collect(),
        }
    }

    pub fn add(&self, rhs: &Self) -> Self {
        let dim = self.dimension().max(rhs.dimension());
        let mut out = vec![0.0; dim];
        for i in 0..dim {
            let a = self.components.get(i).copied().unwrap_or(0.0);
            let b = rhs.components.get(i).copied().unwrap_or(0.0);
            out[i] = a + b;
        }
        Self { components: out }
    }

    pub fn sub(&self, rhs: &Self) -> Self {
        let dim = self.dimension().max(rhs.dimension());
        let mut out = vec![0.0; dim];
        for i in 0..dim {
            let a = self.components.get(i).copied().unwrap_or(0.0);
            let b = rhs.components.get(i).copied().unwrap_or(0.0);
            out[i] = a - b;
        }
        Self { components: out }
    }

    pub fn normalize(&self) -> Result<Self, KernelError> {
        let total = self.magnitude();
        if total == 0.0 {
            return Err(KernelError::ZeroNormalization);
        }
        Ok(self.direction())
    }

    pub fn project_onto(&self, axis: &Self) -> Result<Self, KernelError> {
        let denom = axis.euclidean_norm();
        if denom == 0.0 {
            return Err(KernelError::ZeroNormalization);
        }
        let dot = self
            .components
            .iter()
            .zip(axis.components.iter())
            .map(|(a, b)| a * b)
            .sum::<f64>();
        let scale = dot / (denom * denom);
        Ok(axis.scale(scale))
    }

    pub fn rotate_2d(&self, radians: f64) -> Result<Self, KernelError> {
        if self.dimension() < 2 {
            return Err(KernelError::InvalidOperation("rotate_2d requires at least two components"));
        }
        let (x, y) = (self.components[0], self.components[1]);
        let cos_t = radians.cos();
        let sin_t = radians.sin();
        let mut out = self.components.clone();
        out[0] = x * cos_t - y * sin_t;
        out[1] = x * sin_t + y * cos_t;
        Ok(Self { components: out })
    }

    pub fn constrain(&self, min: &Self, max: &Self) -> Self {
        let dim = self.dimension().max(min.dimension()).max(max.dimension());
        let mut out = vec![0.0; dim];
        for i in 0..dim {
            let v = self.components.get(i).copied().unwrap_or(0.0);
            let lo = min.components.get(i).copied().unwrap_or(f64::MIN);
            let hi = max.components.get(i).copied().unwrap_or(f64::MAX);
            out[i] = v.max(lo).min(hi);
        }
        Self { components: out }
    }

    pub fn nullify(dimensions: usize) -> Self {
        Self::zero(dimensions)
    }
}

#[derive(Clone, Copy, Debug, PartialEq)]
pub struct Percent(pub f64);

impl Percent {
    pub fn new(value: f64) -> Result<Self, KernelError> {
        if !(0.0..=1.0).contains(&value) {
            return Err(KernelError::InvalidStructure("percent out of range"));
        }
        Ok(Self(value))
    }
}

#[derive(Clone, Copy, Debug, PartialEq)]
pub struct AuthRatio(pub f64);

impl AuthRatio {
    pub fn new(value: f64) -> Result<Self, KernelError> {
        if !(0.0..=1.0).contains(&value) {
            return Err(KernelError::InvalidStructure("auth ratio out of range"));
        }
        Ok(Self(value))
    }
}

#[derive(Clone, Copy, Debug, PartialEq)]
pub struct WeightSet {
    pub wm: f64,
    pub wc: f64,
    pub wo: f64,
    pub wp: f64,
}

impl WeightSet {
    pub fn new(wm: f64, wc: f64, wo: f64, wp: f64) -> Result<Self, KernelError> {
        let sum = wm + wc + wo + wp;
        if (sum - 1.0).abs() > 1e-9 {
            return Err(KernelError::InvalidStructure("weights must sum to 1"));
        }
        Ok(Self { wm, wc, wo, wp })
    }
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub enum VectorType {
    Position,
    Free,
    Bound,
    Unit,
    Zero,
}
