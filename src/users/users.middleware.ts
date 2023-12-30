import express from 'express';

import { User } from '@db/entity/User';

import { getUserByEmail, readById } from './users.repository';

const validateRequiredUserBodyFields = (
  req: express.Request,
  res: express.Response,
  next: express.NextFunction,
): void => {
  const body = req.body as User;
  if (body && body.email && body.password) {
    next();
  } else {
    res.status(400).send({
      errors: ['Missing required fields: email and password'],
    });
  }
};

const validateSameEmailDoesntExist = async (
  req: express.Request,
  res: express.Response,
  next: express.NextFunction,
): Promise<void> => {
  const user = await getUserByEmail((req.body as User).email);
  if (user) {
    res.status(400).send({ errors: ['User email already exists'] });
  } else {
    next();
  }
};

const validateSameEmailBelongToSameUser = async (
  req: express.Request,
  res: express.Response,
  next: express.NextFunction,
): Promise<void> => {
  const user = await getUserByEmail((req.body as User).email);
  if (user && user.id === req.params.userId) {
    res.locals.user = user;
    next();
  } else {
    res.status(400).send({ errors: ['Invalid email'] });
  }
};

const validatePatchEmail = async (
  req: express.Request,
  res: express.Response,
  next: express.NextFunction,
): Promise<void> => {
  if ((req.body as User).email) {
    await validateSameEmailBelongToSameUser(req, res, next);
  } else {
    next();
  }
};

const validateUserExists = async (
  req: express.Request,
  res: express.Response,
  next: express.NextFunction,
): Promise<void> => {
  const user = await readById(req.params.userId);
  if (user) {
    next();
  } else {
    res.status(404).send({
      errors: [`User ${req.params.userId} not found`],
    });
  }
};

const extractUserId = (
  req: express.Request,
  res: express.Response,
  next: express.NextFunction,
): void => {
  (req.body as User).id = req.params.userId;
  next();
};

export {
  extractUserId,
  validatePatchEmail,
  validateRequiredUserBodyFields,
  validateSameEmailBelongToSameUser,
  validateSameEmailDoesntExist,
  validateUserExists,
};
