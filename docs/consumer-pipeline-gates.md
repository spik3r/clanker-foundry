# Verify a consumer that uses Foundry

Use three separate gates. They answer different questions and should not share
a single pass label.

| Gate | Run | Evidence |
| --- | --- | --- |
| Model | Validate the reviewed source and project it to named inputs. Run the projector's check mode. | Invalid models are refused and committed projections match the source. |
| Artefacts | Pin a clean Foundry revision in `foundry.json`. Run [`foundrygen plan`, `generate` and `check`](../modeling/foundrygen/README.md). | Each declared output has a matching generator identity and verified manifest. Stale or edited generated files fail the check. |
| Behaviour | Build the consumer and run an independent acceptance runner against its actual public entry point. | Expected and observed results agree. A deliberate fault outside the model is detected. |

Projecting cases from a model can cover all declared routes or states without
duplicating them by hand. It can also duplicate a mistake in the model. Keep
a small set of requirements, expected refusals or public interfaces as
separately reviewed literals. The runner must observe the built application
rather than invoke the same routine that produced the expectation.

When recording a result, name the source and generator revisions, the commands
that ran, the independent cases, what the deliberate fault changed, and
anything that was not exercised. A successful artefact check alone is not
acceptance. None of these steps approves a deploy.

## Worked pilot: Mement0rq article routes

The [public site repository](https://github.com/spik3r/mement0rq-website) has
one site model for route paths, metadata and article status. The pilot extends
its existing architecture and article sequence projections with a generated
browser case inventory, and uses a clean, pinned Foundry checkout for both
diagram bundles. The existing React components are hand-written. This is a
model-driven consumer and an artefact verification example, not a general
application generator.

The browser acceptance runner reads the inventory but not React modules. It
observes the built local SPA: the article index, route links, direct deep links,
headings, metadata, dates, return links and the unlinked draft. Independently
fixed route paths and body headings keep a model mistake from silently
redefining those public expectations.

For the negative control, append an incorrect suffix to the React index
links without editing the model or the case inventory. Building and checking
the model should still pass; browser acceptance should fail on the rendered
hrefs. Apply the fault in an isolated checkout, then remove it and rerun.
A direct HTTP request is insufficient here: an SPA fallback can return the
same shell for both a good route and a nonexistent article.

The pilot covers the site's built client, not Cloudflare's Worker, production
routing, search engine rendering or the whole pipeline for generated code.
For the latter, start with the [bounded lifecycle workflow](model-driven-current-workflow.md)
and its separately authored literal suites.
