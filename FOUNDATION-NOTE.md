# Fabric Foundation Boundary Note

Fabric Foundation is the non-profit R&D and public-good specification layer for the Fabric model.

It defines the shared theory, vocabulary, primitives, axioms, and protocol language required to describe Fabric as a coherent participation system.

Commercial implementation belongs outside the foundation layer.

```text
Fabric Foundation
    = non-profit R&D
    = public-good specification
    = theory and vocabulary
    = primitives and protocols

fabric.db
    = commercial implementation
    = runtime
    = database
    = product
    = service
```

Canonical boundary:

```text
Foundation defines the Fabric.
fabric.db implements the Fabric.
```

## Alignment

The existing repository already contains runnable primitive and runtime work.

This note does not replace that work.

It clarifies the boundary:

- Foundation work remains neutral, public-good, research-oriented, and specification-first.
- Commercial implementation, packaging, pricing, enterprise delivery, and managed services belong in `fabric.db` or other commercial repositories.

## Related Concepts

Fabric Foundation aligns with:

- OpenBox as the opened reality model
- Fabric as the computational model of participation
- Event-driven state
- Drift detection
- Reconciliation
- Recovery
- Coherence over completeness
- Totem-grounded participation
- Agent Access Protocol
- Agent Rules

The foundation should remain useful even when multiple commercial implementations exist.
