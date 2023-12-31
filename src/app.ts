import * as bodyparser from 'body-parser';
import compression from 'compression';
import cors from 'cors';
import express from 'express';
import * as expressWinston from 'express-winston';
import helmet from 'helmet';
import httpStatus from 'http-status';
import passport from 'passport';
import * as winston from 'winston';

import { defaultConfig as config, jwtStrategy, morgan } from '@config/index';
import { ApiError } from '@common/utils/api-error';
import { authLimiter, errorConverter, errorHandler } from '@common/middleware';

import userRoutes from './users/users.routes';
import docRoutes from './docs/docs.routes';

const app: express.Application = express();

app.use(bodyparser.json());
app.use(cors());
app.use(helmet());

const loggerOptions: expressWinston.LoggerOptions = {
  transports: [new winston.transports.Console()],
  format: winston.format.combine(
    winston.format.json(),
    winston.format.prettyPrint(),
    winston.format.colorize({ all: true }),
  ),
};

if (!process.env.DEBUG) {
  loggerOptions.meta = false; // when not debugging, make terse
  if (typeof global.it === 'function') {
    loggerOptions.level = 'http'; // for non-debug test runs, squelch entirely
  }
}

app.use(expressWinston.logger(loggerOptions));

if (config.env !== 'test') {
  app.use(morgan.successHandler);
  app.use(morgan.errorHandler);
}

// set security HTTP headers
app.use(helmet());

// parse json request body
app.use(express.json());

// parse urlencoded request body
app.use(express.urlencoded({ extended: true }));

// gzip compression
app.use(compression());

// enable cors
app.use(cors());
app.options('*', cors());

// jwt authentication
app.use(passport.initialize());
passport.use('jwt', jwtStrategy);

// limit repeated failed requests to auth endpoints
if (config.env === 'production') {
  app.use('/v1/auth', authLimiter);
}

// v1 api routes
app.use('/api/v1', [userRoutes, docRoutes]);

// send back a 404 error for any unknown api request
app.use((req, res, next) => {
  next(new ApiError(httpStatus.NOT_FOUND, 'Not found'));
});

// convert error to ApiError, if needed
app.use(errorConverter);

// handle error
app.use(errorHandler);

export default app;
