import { version } from '../../package.json';
import { defaultConfig as config } from '@config/base';

const swaggerDef = {
  info: {
    title: 'DApps API documentation',
    version,
    license: {
      name: 'MIT',
      url: 'https://github.com/jaganathanb/dapps-api/blob/master/LICENSE',
    },
  },
  host: `localhost:${config.port}`,
  basePath: '/api/v1',
  schemes: ['http', 'https'],
  consumes: ['application/json'],
  produces: ['application/json'],
  securityDefinitions: {
    apiKeyAuth: {
      type: 'Bearer',
      in: 'header',
      name: 'Authoriation',
      description: 'Add jwt token here',
    },
  },
};

export default swaggerDef;
