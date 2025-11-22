# Security Policy

## Supported Versions

Currently supported versions with security updates:

| Version | Supported |
| ------- | --------- |
| 1.0.x   | Yes       |

## Reporting a Vulnerability

If you discover a security vulnerability in gopantic, please report it by:

1. **Opening a GitHub Security Advisory** at https://github.com/vnykmshr/gopantic/security/advisories/new
2. **Filing a private issue** if you prefer not to use Security Advisories

Please include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if available)

### Response Timeline

- Initial response: Within 48 hours
- Status update: Within 7 days
- Fix timeline: Depends on severity (critical issues prioritized)

## Security Considerations

When using gopantic, be aware of:

1. **Error Message Content**: Error messages include field names and values. Avoid logging errors containing sensitive data in production.

2. **Input Size Limits**: Configure `MaxInputSize` when parsing untrusted data to prevent resource exhaustion.

3. **Reflection Usage**: The library uses reflection for type coercion. Ensure your struct definitions don't expose unintended fields.

4. **Dependency Security**: We maintain minimal dependencies (only yaml.v3 in production). Run `go mod verify` regularly and keep dependencies updated.
