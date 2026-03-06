## 2025-05-15 - [RLock for Cache Read Path]
**Learning:** In high-concurrency bots, shared state like caches for chat settings and admin lists are read-heavy. Using a standard `sync.Mutex` for `Get` operations creates a bottleneck by serializing all reads.
**Action:** Always prefer `sync.RWMutex` for caches and use `RLock` for the fast path. Implement lazy eviction using double-checked locking to maintain performance while ensuring consistency.
