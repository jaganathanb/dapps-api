import express from 'express';
import swaggerUi from 'swagger-ui-express';
import sout from '../../swagger-output.json';

const router = express.Router();

router.use('/docs', swaggerUi.serve);
router.get(
  '/docs',
  swaggerUi.setup(sout, {
    explorer: true,
  }),
);

export default router;
