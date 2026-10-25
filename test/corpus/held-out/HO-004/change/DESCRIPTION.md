Make the setup script safe to re-run.

When the API answers 422 to the create call, the user already exists, so go on.

The API's documented answers:

- create: 201 for a new user; 422 for a user that exists; 422 for an input it refuses (an email address it will not take).
  The body of each 422 says which of the two reasons applies.
- team: 204 when the user exists; 404 when it does not.
