package prefixaliases

// byte serializes literal operands for assembly helpers.
func asmByte() {}

// method declares ABI signatures without the `asm` prefix in docs.
func asmMethod() {}

// mimc documents the hash helper that intentionally omits the op prefix.
func opMimc() {}
