# Frontend Rules

@AI.md PART 16, 17: Web Frontend, Admin Panel

## Web Frontend (PART 16)
- Dark theme default (Dracula)
- Mobile-first responsive design
- NO inline CSS (use classes)
- NO inline JavaScript
- WCAG 2.1 AA accessibility

## CSS Rules
- External CSS files only
- No `style=` attributes
- No `<style>` blocks in templates
- Use utility classes

## Admin Panel (PART 17)
- Path: `/{admin_path}` (default: `/admin`)
- Separate from public routes
- Full settings management via WebUI
- Every setting has admin UI

## Template Rules
- Go templates with `.tmpl` extension
- Embedded in binary
- Comments above code only
- Proper escaping for XSS prevention

## Forms
- CSRF tokens required
- Proper validation
- Error messages displayed
