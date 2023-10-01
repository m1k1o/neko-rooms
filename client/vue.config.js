const { defineConfig } = require('@vue/cli-service')
module.exports = defineConfig({
  productionSourceMap: false,
  transpileDependencies: [
    'vuetify'
  ],
  publicPath: './',
  assetsDir: './',
  devServer: {
    allowedHosts: "all",
    proxy: process.env.API_PROXY ? {
      '^/api': {
        target: process.env.API_PROXY,
      },
    } : undefined,
  }
})
