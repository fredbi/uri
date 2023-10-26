v1.1
* [x] investigate - should empty IPv6 be legit (and other nitpicks)? https://url.spec.whatwg.org/#host-parsing. The RFC, the golang ipnet are clear: it is a no
* [x] injects fixes in main (last rune vs once)

v1.2

* [x] strict IP without percent-encoding
* [x] strict UTF8 percent-encoding on host, path, query, userinfo, fragment (not scheme)
* [x] performances on a par with net/url.URL
* [x] reduce # allocs String() ? (~ predict size)
* [x] more nitpicks - check if the length checks on DNS host name are in bytes or in runes => bytes
* [x] DefaultPort(), IsDefaultPort()
* [x] fix scheme tolerance to be on ASCII only
* [x] document the choice of a strict % escaping policy regarding char encoding (UTF8 mandatory)
* [] handle empty fragment, empty query
* [] fix scheme tolerance to be on ASCII only
* [] IRI ucs charset compliance (att: perf challenge)
* [] Support IRI iprivate in query
* [] normalizer

v2.0.0

* [x] V2 zero alloc, no interface, fluent builder with inner error checking
* [] doc: complete the librarian/archivist work on specs, etc + FAQ to clarify the somewhat arcane world of this set of RFCs.
* [] V2 net/url.URL interface-feeling (more methods)

Future/Maybe

* [] simplify: readability vs performance
* [] investigate - nice? uses whatwg errors terminology

### Notes
```
ucschar        = %xA0-D7FF / %xF900-FDCF / %xFDF0-FFEF
                  / %x10000-1FFFD / %x20000-2FFFD / %x30000-3FFFD
                  / %x40000-4FFFD / %x50000-5FFFD / %x60000-6FFFD
                  / %x70000-7FFFD / %x80000-8FFFD / %x90000-9FFFD
                  / %xA0000-AFFFD / %xB0000-BFFFD / %xC0000-CFFFD
                  / %xD0000-DFFFD / %xE1000-EFFFD
```

```
iprivate       = %xE000-F8FF / %xF0000-FFFFD / %x100000-10FFFD

		// TODO: RFC6874
		//  A <zone_id> SHOULD contain only ASCII characters classified as
   		// "unreserved" for use in URIs [RFC3986].  This excludes characters
   		// such as "]" or even "%" that would complicate parsing.  However, the
   		// syntax described below does allow such characters to be percent-
   		// encoded, for compatibility with existing devices that use them.
```
