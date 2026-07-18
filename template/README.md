## Template Engine 

**Preparation**

Place your base HTML files and JSON object files in the same directory:
```text
templates/
├── en/
├── ── layouts.html
├── ── components.json
└── ── email.html
├── uk/
├── ── layouts.html
├── ── components.json
└── ── email.html
```

**Object File Format**

Object files are JSON maps where keys represent the snippet name and values contain the HTML string.

Example: components.json

```json
{
  "header_alert": "<div class='alert'>IMPORTANT: {message}</div>",
  "action_button": "<a href='{url}' class='btn'>{label}</a>",
  "divider": "<hr style='border: 1px solid #eee;' />"
}
```

**Usage**

1. Initialize and Load

Create a Store to cache templates and a Templator to process them.

```go
store := template.NewStore()
err := store.LoadFromDir("./templates")
if err != nil {
    log.Fatal(err)
}

templ := template.NewTemplator(store)
```

2. Format a Template

Define a Trequest with the base template name, required objects, and data for replacement.

```go
req := template.Trequest{
    BaseName: "layouts",
    Objects: []string{
        "components.header_alert",
        "components.action_button",
    },
    Data: []map[string]string{
        {"{message}": "Subscription expiring"},
        {"{url}": "https://example.com/pay"},
        {"{label}": "Renew Now"},
    },
}

html, err := templ.Format(req)
if err != nil {
    log.Printf("Error: %v", err)
}
```

Core Logic

    {body}: A reserved placeholder in your base HTML files. All snippets requested in the Objects slice are concatenated and injected into this tag.

    Replacement: All keys provided in the Data maps are replaced in a single pass across the entire assembled document.

    Redaction: If no Objects are provided, the {body} tag is automatically removed from the output.