<!-- BEGIN:nextjs-agent-rules -->
# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` before writing any code. Heed deprecation notices.
<!-- END:nextjs-agent-rules -->

# Components must have Storybook stories

**Every new component in `site/src/components/` MUST ship with a matching `.stories.tsx` file in the same directory.** No exceptions — if you create `foo.tsx`, you also create `foo.stories.tsx` as part of the same change.

Stories must cover:
- The default / most common state
- Every meaningful variant (size, severity, status, disabled, loading, etc.)
- Edge cases that affect layout (empty, overflowing, long text, many items)
- Any interactive state worth visualizing (expanded, selected, hovered, error)

This rule applies when:
- Creating a brand new component
- Extracting a sub-component from an existing one
- Porting a component from another library or codebase
- Building composed/page-level components (not just atoms)

The reasoning: Storybook is Dispatch's primary frontend development environment. Components without stories can't be visually inspected in isolation, can't have variants exercised, and tend to drift out of sync with the CVA variant system. A component that only exists inside `page.tsx` might break in ways nobody notices until production. Stories force you to think through states explicitly.

This also applies when modifying an existing component — if you add a new variant or prop, add a corresponding story for it.
