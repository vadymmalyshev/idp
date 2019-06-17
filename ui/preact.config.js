// preact.config.js
import webpack from 'webpack';
import asyncPlugin from 'preact-cli-plugin-async';
import dotenv from 'dotenv';

// module.exports = function (config) {
//   config.node.process = true;

//   // config.plugins.push(
//   //   new webpack.DefinePlugin({
//   //     BASE_PATH: JSON.stringify('https://id.hiveon.net'),
//   //   })
//   // );

  
// }

export default (config, env, helpers) => {
  config.node.process = true;
  const envConf = dotenv.config();
  console.log('envConf: ', envConf);
  config.devServer = {
    quiet: false,
    clientLogLevel: 'info',
    proxy: [
      {
        path: '/api/**',
        changeOrigin: true,
        changeHost: true,
        target: process.env.BASE_PATH || 'https://id.hiveon.dev',
        // ...any other stuff...
        // optionally mutate request before proxying:
        pathRewrite: function(path, req) {
          // you can modify the outbound proxy request here:
          delete req.headers.referer;
          // console.log('ssss', path)

          // common: remove first path segment: (/api/**)
          // return '/' + path.replace(/^\/[^\/]+\//, '');
        },

        // optionally mutate proxy response:
        // onProxyRes: function(proxyRes, req, res) {
        //   // you can modify the response here:
        //   proxyRes.headers.connection = 'keep-alive';
        //   proxyRes.headers['cache-control'] = 'no-cache';
        // },
        onProxyRes: proxyRes => {
          // Disable browser caching
          /* eslint-disable no-param-reassign, dot-notation */
          proxyRes.headers['Cache-Control'] = 'no-cache, no-store, must-revalidate';
          proxyRes.headers['Pragma'] = 'no-cache';
          proxyRes.headers['Expires'] = 0;
        },
        // onError: (err, req, res) => {
        // console.error(err);
        // if(res.sentry){
        //   res.statusCode = 500;
        //   res.end(res.sentry + '\n');
        // }
        // },
        onProxyReq: (proxyReq, req) => {
          const ip = req.headers['x-forwarded-for'] || req.connection.remoteAddress;
          // console.log(proxyReq)
          if (req.body) {
            const bodyData = JSON.stringify(req.body);
            proxyReq.setHeader('X-Forwarded-For', ip.replace('::ffff:', ''));
            proxyReq.setHeader('Content-Type', 'application/json');
            proxyReq.write(bodyData);
            proxyReq.end();
          }
        },
      }

    ]
  }
  if (!env.production) {
    config.devServer.public = 'id.hiveon.dev';
    config.devServer.stats = 'minimal';
  }
  asyncPlugin(config);
}