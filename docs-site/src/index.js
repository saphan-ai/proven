// The Worker in front of the docs container: every request goes to the
// container's port. A static docs site needs no secret and no storage.
import { DurableObject } from 'cloudflare:workers';

const PORT = 3000;

export class DocsContainer extends DurableObject {
  constructor(ctx, env) {
    super(ctx, env);
    // no outbound network: the site only serves pages
    this.ctx.blockConcurrencyWhile(async () => {
      if (!this.ctx.container.running) this.ctx.container.start({ enableInternet: false });
    });
  }

  async fetch(request) {
    const port = this.ctx.container.getTcpPort(PORT);
    // the container speaks plain HTTP; the public side is already TLS
    const url = new URL(request.url);
    url.protocol = 'http:';
    url.port = '';
    // the server needs a moment after a cold start; try for about 20 seconds
    let last;
    for (let attempt = 0; attempt < 20; attempt++) {
      try {
        return await port.fetch(new Request(url, request.clone()));
      } catch (err) {
        last = err;
        await new Promise((resolve) => setTimeout(resolve, 1000));
      }
    }
    return new Response(`the documentation is starting; try again shortly (${last})`, { status: 503, headers: { 'retry-after': '5' } });
  }
}

export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    // DOCS_HOST is the one public host; any other name is sent there
    if (env.DOCS_HOST && url.hostname !== env.DOCS_HOST) {
      url.hostname = env.DOCS_HOST;
      url.protocol = 'https:';
      url.port = '';
      return Response.redirect(url.toString(), 301);
    }
    const stub = env.DOCS_CONTAINER.get(env.DOCS_CONTAINER.idFromName('docs'));
    return stub.fetch(request);
  },
};
