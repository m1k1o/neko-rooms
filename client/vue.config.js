module.exports = {
  productionSourceMap: false,
  transpileDependencies: [
    'vuetify'
  ],
  publicPath: './',
  assetsDir: './',
  devServer: {
    disableHostCheck: true,
    proxy: process.env.API_PROXY ? {
      '^/api': {
        target: process.env.API_PROXY,
      },
    } : undefined,
  }
}
