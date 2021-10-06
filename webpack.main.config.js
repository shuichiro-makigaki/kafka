const CopyWebpackPlugin = require("copy-webpack-plugin");
const path = require("path");

module.exports = {
  /**
   * This is the main entry point for your application, it's the first file
   * that runs in the main process.
   */
  entry: './src/index.ts',
  // Put your normal webpack config below here
  module: {
    rules: require('./webpack.rules'),
  },
  plugins: [
    new CopyWebpackPlugin({
      patterns: [
        {from: path.resolve(__dirname, 'kafka-server', 'kafka-server.exe'), to: 'kafka-server.exe'},
        {from: path.resolve(__dirname, 'kafka-server', 'ffmpeg-4.4-full_build'), to: 'ffmpeg-4.4-full_build/'}
      ]
    })
  ],
  resolve: {
    extensions: ['.js', '.ts', '.jsx', '.tsx', '.css', '.json']
  }
};