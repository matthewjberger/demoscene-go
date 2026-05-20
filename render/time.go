package render

// DeltaTime is the duration of the most recent frame, in seconds.
// Stored as a typed resource on the engine world (and the game world,
// when one is used). Platform layers stamp it once per frame before
// running the schedule; systems read it from the world.
type DeltaTime float32
