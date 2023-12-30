const allRoles = {
  user: [] as string[],
  admin: ['getUsers', 'manageUsers'] as string[],
};

const roles = Object.keys(allRoles);
const roleRights = new Map(Object.entries(allRoles));

export { roles, roleRights };
