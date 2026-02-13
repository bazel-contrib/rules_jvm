# Class resolve over regexp

Test that class-level `gazelle:resolve` directives take precedence over
package-level `gazelle:resolve_regexp` directives.

When a class has an explicit `resolve` directive (e.g., for 
`com.google.common.util.concurrent.RateLimiterCreator`), it should be used
instead of a broader `resolve_regexp` that matches the package (e.g.,
`com\.google\.common\..*`).
