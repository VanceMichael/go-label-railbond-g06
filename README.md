# RailBond G06 Annotation Root

This public repository carries the independent Go annotation tasks for the RailBond G06 bonded cross-border rail operations backend.

Each task is isolated under `tasks/<task_key>/red` and `tasks/<task_key>/green`.

- `green` starts at a parentless G1 commit containing one runtime defect and no task-private test.
- `red` is a parentless R1 commit containing the same G1 tree plus that task's private test-only patch.
- Task-private tests and authoring evidence are not part of the model-visible baseline.
