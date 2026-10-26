import { createMDX } from 'fumadocs-mdx/next';

const withMDX = createMDX();

/** @type {import('next').NextConfig} */
const config = {
  reactStrictMode: true,
  output: 'standalone',
  // the image copies .next/standalone, so server.js must sit at its top
  outputFileTracingRoot: import.meta.dirname,
};

export default withMDX(config);
