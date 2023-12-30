import express, { Response } from 'express';
import { validationResult } from 'express-validator';

const verifyBodyFieldsErrors = (
  req: express.Request,
  res: express.Response,
  next: express.NextFunction,
): void | Response => {
  const errors = validationResult(req);
  if (!errors.isEmpty()) {
    return res.status(400).send({ errors: errors.array() });
  }
  next();
};

export { verifyBodyFieldsErrors };
