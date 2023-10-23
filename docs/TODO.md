* []  investigate - should empty IPv6 be legit (and other nitpicks)? https://url.spec.whatwg.org/#host-parsing
* []  investigate - nice? uses whatwg errors terminology
* [x] injects fixes in main (last rune vs once)
* []  doc: complete the librarian/archivist work on specs, etc + FAQ to clarify the somewhat arcane world of this set of RFCs.
* [x] strict IP without percent-encoding
* [x] strict UTF8 percent-encoding on host, path, query, userinfo, fragment (not scheme)
* [] simplify
* [x] performances on a par with net/url.URL
* [] reduce # allocs String() ? (~ predict size)
* [x] more nitpicks - check if the length checks on DNS host name are in bytes or in runes => bytes
* [] IRI ucs charset compliance
* [] DefaultPort(), IsDefaultPort()
* [] normalizer
