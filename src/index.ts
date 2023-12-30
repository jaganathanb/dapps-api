import http, { IncomingMessage, ServerResponse } from 'http';

import { AppDataSource } from '@db/data-source';
import { logger } from '@config/logger';
import { defaultConfig as config } from '@config/base';

import app from './app';

let server: http.Server<typeof IncomingMessage, typeof ServerResponse>;
AppDataSource.initialize()
  .then(() => {
    logger.info('Data Source has been initialized!');

    server = app.listen(config.port, () => {
      logger.info(`Listening to port ${config.port}`);
    });
  })
  .catch((err) => {
    logger.error('Error during Data Source initialization:', err);
  });

const exitHandler = () => {
  if (server) {
    server.close(() => {
      logger.info('Server closed');
      process.exit(1);
    });
  } else {
    process.exit(1);
  }
};

const unexpectedErrorHandler = (error: unknown) => {
  logger.error(error);
  exitHandler();
};

process.on('uncaughtException', unexpectedErrorHandler);
process.on('unhandledRejection', unexpectedErrorHandler);

process.on('SIGTERM', () => {
  logger.info('SIGTERM received');
  if (server) {
    server.close();
  }
});
