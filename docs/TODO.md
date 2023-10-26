* [x] investigate - should empty IPv6 be legit (and other nitpicks)? https://url.spec.whatwg.org/#host-parsing. The RFC, the golang ipnet are clear: it is a no
* [x] injects fixes in main (last rune vs once)
* [x] strict IP without percent-encoding
* [x] strict UTF8 percent-encoding on host, path, query, userinfo, fragment (not scheme)
* [x] performances on a par with net/url.URL
* [x] reduce # allocs String() ? (~ predict size)
* [x] more nitpicks - check if the length checks on DNS host name are in bytes or in runes => bytes
* [x] DefaultPort(), IsDefaultPort()
* [] IRI ucs charset compliance (att: perf challenge)
* [] normalizer
* [] V2 zero alloc, no interface, fluent builder with inner error checking
* [] doc: complete the librarian/archivist work on specs, etc + FAQ to clarify the somewhat arcane world of this set of RFCs.
* [] simplify: readability vs performance
* [] investigate - nice? uses whatwg errors terminology
* [] document the choice of a strict % escaping policy regarding char encoding (UTF8 mandatory)
