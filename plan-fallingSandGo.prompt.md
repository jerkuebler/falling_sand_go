## Plan: Scale and Enrich Falling Sand

Improve the simulation in three layers: first add measurement and correctness guardrails, then reduce per-frame work in the world update path so particle count scales with active regions instead of screen size, then expand the material model so richer reactions can be added without turning the simulation loop into a large special-case dispatcher.

**Steps**
1. Phase 1: establish a performance baseline before changing architecture. Reuse the existing benchmark in [internal/world/world_test.go](internal/world/world_test.go) and extend it with representative densities and grid sizes so later changes can be compared cleanly. Add focused correctness tests for movement, edge handling, and representative reactions.
2. Phase 1: instrument the hot path in [internal/world/world.go](internal/world/world.go), especially `World.UpdateWorld`, `directionalNodeCheck`, `holdOrDisplace`, `nearestBlank`, and the draw/update boundary. The point is to confirm how much time is going to full-grid iteration, buffer copy/clear, neighbor checks, and randomness.
3. Phase 2: reduce work from whole-grid scanning to active-region scanning. Replace the unconditional nested loop in `World.UpdateWorld` with chunk-based occupancy tracking or an active-particle list. Recommended first move: fixed-size chunks with active/dirty tracking, because it preserves the current grid-based rules and avoids a full sparse-particle rewrite.
4. Phase 2: remove the full-frame `copy` and `clear` cost in `World.UpdateWorld`. Move to buffer swapping plus explicit tracking of written cells or active chunks. This depends on step 3, because once activity is tracked, only touched regions need resetting.
5. Phase 2: flatten hot-path dispatch and metadata lookups. The current `nodeFuncs` and `MaterialInteractions` maps are readable, but map lookups in the per-particle inner loop are unnecessary overhead. Convert hot-path lookups to array/table access indexed by `NodeType`.
6. Phase 2: reduce random and branch overhead in movement helpers in [internal/utils/utils.go](internal/utils/utils.go). Rework `RandomLateral`, `RandomDownDiagonal`, `RandomUpDiagonal`, and similar helpers to use cheaper alternation or precomputed patterns. This can run in parallel with step 5.
7. Phase 2: redesign `Node` storage in [internal/material/material.go](internal/material/material.go) for better cache density. Consider packing `NodeType` and flags more tightly or moving transient frame state like `Dirty` out of the primary material buffer. This depends on the chosen update architecture from steps 3 and 4.
8. Phase 3: separate movement behavior from reaction/state behavior. Keep movement in [internal/world/world.go](internal/world/world.go), but move reactions into a material/rules layer in [internal/material/material.go](internal/material/material.go) so new content is data-driven instead of branching through world logic.
9. Phase 3: add a lightweight per-particle state model. Recommended starting fields: temperature and lifetime. Temperature is the best first addition because the existing lava/water/steam system already implies it, and it unlocks boiling, condensation, cooling, ignition, and melting cleanly.
10. Phase 3: add new interactions in increasing complexity order. Recommended order: rebalance water/lava/rock/steam transitions, improve condensation, add wet sand or sludge-like behavior, then add fire/smoke/ash, then oil or acid families, then pressure or explosion effects.
11. Phase 3: revisit rendering only after simulation scaling improves. [pkg/main.go](pkg/main.go) currently writes every pixel every frame. If chunking makes simulation sparse, consider dirty-rectangle redraws or simulation/render decoupling, but only if profiling shows draw cost has become material.
12. Phase 4: evaluate parallel chunk updates if single-thread improvements are not enough. Do not parallelize the current loop directly; the present `Dirty` and `next` semantics are too tightly coupled. Concurrency should come only after chunk ownership and cross-chunk write rules are explicit.

**Relevant files**
- [internal/world/world.go](internal/world/world.go) — primary hot path; likely refactor `World.UpdateWorld`, `updateFuncs`, `directionalNodeCheck`, `holdOrDisplace`, `nearestBlank`, and buffer management.
- [internal/world/world_test.go](internal/world/world_test.go) — expand `BenchmarkUpdate` into a real performance and regression suite.
- [internal/material/material.go](internal/material/material.go) — centralize material definitions and evolve `Node`, `nodeInfo`, and `MaterialInteractions` into a table-driven system.
- [internal/utils/utils.go](internal/utils/utils.go) — reduce randomness overhead in helpers used from the simulation inner loop.
- [pkg/main.go](pkg/main.go) — useful for timing overlays, active-particle counters, and future simulation/render decoupling experiments.

**Verification**
1. Add benchmark cases for multiple grid sizes and occupancy levels, then compare `ns/op`, `allocs/op`, and effective TPS before and after each optimization phase.
2. Add focused tests for top-edge gas behavior, bottom-edge liquid behavior, displacement, and at least two representative reactions.
3. For each major optimization step, run `go test ./...` and world benchmarks with fixed seeds so correctness and performance regressions are attributable.
4. Add a temporary debug overlay showing active particle count, active chunks, update time, and draw time to verify the actual bottleneck is moving in the right direction.
5. Stress-test with a larger world than the current default and verify both stability and interaction quality under sustained load.

**Decisions**
- Recommended architecture direction: chunked active-region simulation first, not a full sparse-particle rewrite.
- Recommended interaction direction: add a generic reaction/state system before adding many new materials.
- Included scope: runtime efficiency, particle-capacity scaling, and extensible interaction design.
- Excluded scope: visual polish, UI/UX work, audio, save/load, and networking.

**Further Considerations**
1. Choose chunking over a sparse particle list unless profiling shows the world is usually extremely sparse.
2. Keep new per-particle state minimal at first; temperature plus lifetime is enough to unlock much richer behavior.
3. Only pursue concurrency after single-thread hot-path cleanup and chunk ownership are in place.
