# Generic Package Exclusion Rules Extraction Tool

This tool extracts generic package exclusion rules from the step-report-json section of the Jenkins log file and formats them for YAML configuration.

## Files

- `main.go` - Minimal Go program
- `template.tmpl` - YAML template

## Usage

### Direct Usage

```bash
go run main.go <log-file> <template-file> > output.yaml
```

### Example

```bash
# Extract generic package exclusions
go run main.go conforma-openshift-jenkins-rhel8-wzhb8-verify.log template.tmpl > generic_exclusions.yaml
```

## Customizing the Template

Edit `template.tmpl` to customize the output format and values:

```yaml
volatileConfig:
  exclude:
{{range .}}    - value: {{.}}
      effectiveUntil: "2025-12-31T00:00:00Z"
      reference: https://issues.redhat.com/browse/PSX-908
{{end}}
```

Change `effectiveUntil` and `reference` to your values, then re-run the command.

## Output

The tool extracts generic package exclusion rules:

- `sbom_spdx.allowed_package_sources:pkg:generic/...` - Jenkins plugins and other packages

Each rule is formatted with the configured `effectiveUntil` date and `reference` URL from the template.

## How It Works

1. Scans the log file for `step-report-json` section
2. Parses the multi-line JSON structure
3. Recursively extracts all strings matching "To exclude this rule add" pattern
4. Deduplicates and sorts the exclusion rules
5. Applies your template to format the output

