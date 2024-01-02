import swaggerAutogen from 'swagger-autogen';
import swaggerDef from './swagger.def';

const outputFile = '../../swagger-output.json';
const endpointsFiles = ['./src/users/users.routes.ts'];

swaggerAutogen({ openapi: '3.0.0' })(
  outputFile,
  endpointsFiles,
  swaggerDef,
).then(async () => {
  await import('../index'); // Your project's root file
});
