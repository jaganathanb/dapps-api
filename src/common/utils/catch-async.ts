import { NextFunction } from 'express';

export const catchAsync =
  (fn: CallableFunction) =>
  (req: Request, res: Response, next: NextFunction): void => {
    Promise.resolve(fn(req, res, next)).catch((err) => next(err));
  };
