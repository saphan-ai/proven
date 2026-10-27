export const metadata = {
  title: 'The fleet in its own words',
  description: 'The five masters, the scribe and the advisor that built Proven answer six questions about their week.',
};

export default function Fleet() {
  return (
    <main className="mx-auto w-full max-w-4xl px-4 py-12">
      <h1 className="mb-4 text-3xl font-bold">The fleet in its own words</h1>
      <p className="mb-8 text-fd-muted-foreground">
        The five masters, the scribe and the advisor that built Proven answer six questions about their week.
      </p>
      <div className="relative w-full overflow-hidden rounded-lg" style={{ paddingTop: '56.25%' }}>
        <iframe
          title="vimeo-player"
          src="https://player.vimeo.com/video/1239909255?h=a9ae236b92"
          className="absolute left-0 top-0 h-full w-full"
          frameBorder="0"
          referrerPolicy="strict-origin-when-cross-origin"
          allow="autoplay; fullscreen; picture-in-picture; clipboard-write; encrypted-media; web-share"
          allowFullScreen
        />
      </div>
      <p className="mt-6">
        Also on Vimeo: <a className="underline" href="https://vimeo.com/user264829523/fleet">vimeo.com/user264829523/fleet</a>
      </p>
      <p className="mt-2">
        <a className="underline" href="/docs">The documentation</a>
      </p>
    </main>
  );
}
