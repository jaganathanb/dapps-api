import httpStatus from 'http-status';
import { NextFunction, Request, Response } from 'express';

import { defaultConfig as config, logger } from '@config/index';

import { ApiError } from '../utils/api-error';

const errorConverter = (
  err: Error,
  req: Request,
  res: Response,
  next: NextFunction,
): void => {
  let error: Error = err;
  if (!(error instanceof ApiError)) {
    const statusCode = httpStatus.INTERNAL_SERVER_ERROR;
    const message = error.message || httpStatus[statusCode];
    error = new ApiError(statusCode, message, false, err.stack);
  }
  next(error);
};

const errorHandler = (err: ApiError, req: Request, res: Response): void => {
  let { statusCode, message } = err;
  if (config.env === 'production' && !err.isOperational) {
    statusCode = httpStatus.INTERNAL_SERVER_ERROR;
    message = httpStatus[httpStatus.INTERNAL_SERVER_ERROR];
  }

  res.locals.errorMessage = err.message;

  const response = {
    code: statusCode,
    message,
    ...(config.env === 'development' && { stack: err.stack }),
  };

  if (config.env === 'development') {
    logger.error(err);
  }

  res.status(statusCode).send(response);
};

export { errorConverter, errorHandler };
