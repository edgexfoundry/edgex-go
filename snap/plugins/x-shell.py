import os

import snapcraft

class ShellPlugin(snapcraft.BasePlugin):
	@classmethod
	def schema(cls):
		schema = super().schema()

		schema['required'] = []

		schema['properties']['shell'] = {
			'type': 'string',
			'default': '/bin/sh',
		}
		schema['required'].append('shell')

		schema['properties']['shell-flags'] = {
			'type': 'array',
			'items': {
				'type': 'string',
			},
			'default': [],
		}

		schema['properties']['shell-command'] = {
			'type': 'string',
		}
		schema['required'].append('shell-command')

		return schema

	def env(self, root):
		return super().env(root) + [
			'SNAPDIR=' + os.getcwd(),
		]

	def build(self):
		super().build()

		return self.run([
			self.options.shell,
		] + self.options.shell_flags + [
			'-c', self.options.shell_command,
		])

# vim:set ts=4 noet:
