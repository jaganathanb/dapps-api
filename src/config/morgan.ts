import morgan from 'morgan';
import { Response } from 'express';

import { defaultConfig as config } from './base';
import { logger } from './logger';

morgan.token(
  'message',
  (req, res: Response) => (res.locals.errorMessage as string) || '',
);

const getIpFormat = () =>
  config.env === 'production' ? ':remote-addr - ' : '';
const successResponseFormat = `${getIpFormat()}:method :url :status - :response-time ms`;
const errorResponseFormat = `${getIpFormat()}:method :url :status - :response-time ms - message: :message`;

const successHandler = morgan(successResponseFormat, {
  skip: (req, res) => res.statusCode >= 400,
  stream: { write: (message) => logger.info(message.trim()) },
});

const errorHandler = morgan(errorResponseFormat, {
  skip: (req, res) => res.statusCode < 400,
  stream: { write: (message) => logger.error(message.trim()) },
});

export { successHandler, errorHandler };
