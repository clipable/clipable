/** @type {import('next').NextConfig} */
const nextConfig = {
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: `http://localhost:8080/api/:path*`,
      }
    ]
  },
  output: 'standalone',
  experimental: {
    appDir: true,
  }
}

module.exports = nextConfig