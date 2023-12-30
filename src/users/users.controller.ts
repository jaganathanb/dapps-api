import argon2 from 'argon2';
import debug from 'debug';
import { Request, Response } from 'express';

import type { User } from '@db/entity/User';

import {
  create,
  deleteById,
  list,
  patchById,
  putById,
  readById,
} from './users.repository';

const log: debug.IDebugger = debug('app:users-controller');

const listUsers = async (req: Request, res: Response): Promise<void> => {
  const users = await list(100, 0);

  res.status(200).send(users);
};

const getUserById = async (req: Request, res: Response): Promise<void> => {
  const user = await readById(req.params.userId);
  res.status(200).send(user);
};

const createUser = async (req: Request, res: Response): Promise<void> => {
  const body = req.body as User;
  body.password = await argon2.hash(body.password);
  const userId = await create(body);
  res.status(201).send({ id: userId });
};

const patch = async (req: Request, res: Response): Promise<void> => {
  const body = req.body as User;
  if (body.password) {
    body.password = await argon2.hash(body.password);
  }
  log(await patchById(req.params.userId, body));
  res.status(204).send();
};

const put = async (req: Request, res: Response): Promise<void> => {
  const body = req.body as User;
  body.password = await argon2.hash(body.password);
  log(await putById(req.params.userId, body));
  res.status(204).send();
};

const removeUser = async (req: Request, res: Response): Promise<void> => {
  log(await deleteById(req.params.userId));
  res.status(204).send();
};

export { createUser, getUserById, listUsers, patch, put, removeUser };
