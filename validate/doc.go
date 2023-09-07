// Package validate is used to ensure that the renaming operation cannot result
// in conflicts before the operation is carried out. It protects against the
// following scenarios:
//
// 1. Overwriting a newly renamed path.
// 2. Target destination contains forbidden characters (varies based on the operating system).
// 3. Target destination already exists on the file system (except if
// --allow-overwrite is specified)
// 4. Target name exceeds the maximum allowed length (255 characters in windows, and 255 bytes on Linux and macOS).
// 5. Target destination contains trailing periods in any of the sub paths (Windows only).
// 6. Target destination is empty.
//
// It detects each conflicts and reports them, but it can also automatically fix
// them according to predefined rules (if -F/--fix-conflicts is specified).
package validate
