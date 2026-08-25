# Swift Standards

Swift-specific standards. The general coding standards in `~/.agents/sdlc/CODING.md` apply on top of these.

Primary references:

- [Swift API Design Guidelines](https://swift.org/documentation/api-design-guidelines/)
- [The Swift Programming Language](https://docs.swift.org/swift-book/documentation/the-swift-programming-language/)
- [swift-format](https://github.com/swiftlang/swift-format)
- [Apple Human Interface Guidelines](https://developer.apple.com/design/human-interface-guidelines) for app/UI work

## Tooling

Use the project's existing configuration first.

| Tool | Purpose | Required |
|---|---|---|
| `swift-format` | Formatting | Yes, unless the project already standardizes on another formatter |
| SwiftLint | Linting | Yes, when the project has or needs lint configuration |
| XCTest / Swift Testing | Tests | Yes, use whichever the project already uses |

Run formatting and linting on changed Swift files before review. Do not introduce a new formatter or lint rule set into an existing project without explicit reason.

## API Design

Follow the Swift API Design Guidelines.

- Optimize for clarity at the call site.
- Prefer names that read naturally where used, not only where declared.
- Avoid abbreviations unless they are established domain terms.
- Use argument labels deliberately; they are part of the API.
- Do not add "Manager", "Helper", or "Util" types unless the responsibility is genuinely coherent.

## Optionals

Avoid force unwrapping.

```swift
// BAD
let image = UIImage(named: name)!

// GOOD
guard let image = UIImage(named: name) else {
    throw ImageError.missing(name)
}
```

Allowed exceptions require a nearby comment naming the invariant that makes the force unwrap safe, usually in tests or static fixtures only.

Use `guard` for early exits when it reduces nesting. Prefer `if let` when the optional branch is local and short.

## Error Handling

- Never use empty `catch`.
- Avoid `try!`; use `try`, `try?`, or propagate the error.
- Preserve context when mapping errors.
- Do not collapse recoverable errors into booleans unless the caller truly does not need to know why it failed.

```swift
// BAD
do {
    try save()
} catch {
}

// GOOD
do {
    try save()
} catch {
    logger.error("Saving document failed: \(error)")
    throw error
}
```

## Concurrency

Prefer structured concurrency.

- Use `async`/`await` for asynchronous flows.
- Avoid `Task.detached` unless isolation from the current task is required and documented.
- Keep UI state on the main actor.
- Do not update SwiftUI/AppKit/UIKit state from background tasks.
- Cancellation must be considered for long-running tasks.

## SwiftUI and Apple UI Code

For user-facing UI, Apple Human Interface Guidelines apply.

- Views should describe UI, not contain business logic.
- Keep side effects out of `body`.
- Use explicit state ownership: `@State`, `@Binding`, `@StateObject`, `@ObservedObject`, and environment values deliberately.
- Do not hide networking, persistence, or process execution inside views.
- Extract view models or service types when behaviour needs testing outside the UI.

## Project Layout

Follow the existing project layout. For new Swift Package Manager projects:

```text
Package.swift
Sources/<ModuleName>/
Tests/<ModuleName>Tests/
```

For app projects, follow Xcode's structure unless there is a clear project convention. Do not reorganize an Xcode project merely because files are being touched.

## Testing

See `~/.agents/sdlc/TESTING.md` for cross-language testing standards.

- Test user-observable behaviour, not private implementation details.
- Use async tests for async APIs.
- Keep test names behaviour-focused.
- UI tests are appropriate when the user-observable result cannot be verified below the UI layer.
- Snapshot tests require explicit justification; they are brittle when used as a substitute for behavioural assertions.

## Dependencies

Prefer Foundation, Swift standard library, and platform frameworks before adding packages.

Any new package must be justified by:

- maintenance status
- platform support
- why standard libraries/frameworks are insufficient
- impact on build and release

## Interop

Objective-C and C interop should be isolated behind small Swift APIs. Do not let pointer-heavy or Objective-C-shaped APIs leak through the application unless the project is specifically an interop layer.

# Canary
Suffix the canary string with "SWIFT "
