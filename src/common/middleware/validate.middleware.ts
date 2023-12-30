import { NextFunction, Request, Response } from 'express';
import httpStatus from 'http-status';
import Joi, { Schema } from 'joi';
import { pick } from 'lodash';

import { ApiError } from '../utils/api-error';

const validate =
  (schema: Schema) =>
  (req: Request, res: Response, next: NextFunction): void => {
    const validSchema = pick(schema, ['params', 'query', 'body']);
    const object = pick(req, Object.keys(validSchema));
    const result = Joi.compile(validSchema)
      .prefs({ errors: { label: 'key' }, abortEarly: false })
      .validate(object);

    if (result.error) {
      const errorMessage = result.error.details
        .map((details) => details.message)
        .join(', ');

      return next(new ApiError(httpStatus.BAD_REQUEST, errorMessage));
    }
    Object.assign(req, result.value);

    return next();
  };

export { validate };
