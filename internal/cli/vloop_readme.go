package cli

import "fmt"

func vloopReadme(lang, version string) string {
	if lang == "es" {
		return fmt.Sprintf(`# vloop

Escrito por vloop %s. Este archivo es de vloop: `+"`vloop upgrade`"+` lo reescribe, así que lo que se edite aquí se pierde.

## Qué es vloop

vloop es una herramienta de línea de comandos y un plugin para Claude Code y GitHub Copilot CLI. Convierte un brief en tareas y, para cada una, una sesión de trabajo la realiza, una compuerta la comprueba y una revisión independiente la juzga, con un commit por tarea. Funciona en macOS, Linux y Windows.

## Cómo se usa

Instálalo con `+"`go install github.com/mvelosop/vloop/cmd/vloop@latest`"+`. Ejecuta `+"`vloop doctor`"+` para comprobar la configuración. Escribe un brief con `+"`/vloop:create-brief`"+` en el agente y ejecútalo con `+"`vloop run <brief>`"+`. Mantén vloop al día con `+"`vloop self-update`"+` y después `+"`vloop upgrade`"+` en cada repositorio.

## Qué hay en esta carpeta

- `+"`config.toml`"+`: configuración compartida.
- `+"`config.local.toml`"+`: configuración de esta máquina, ignorada por Git.
- `+"`install.json`"+`: datos de la instalación de vloop.
- `+"`repo-guidelines.md`"+`: lecciones del repositorio.
- `+"`README.md`"+`: esta descripción, generada por vloop.
- `+"`defects/`"+`: defectos registrados.
- `+"`interventions/`"+`: intervenciones registradas.
- `+"`state/`"+`: el plan, las ejecuciones, las métricas y las revisiones de briefs.
- `+"`tmp/`"+`: archivos temporales, ignorados por Git.

## Más

https://github.com/mvelosop/vloop
`, version)
	}
	return fmt.Sprintf(`# vloop

Written by vloop %s. vloop owns this file: `+"`vloop upgrade`"+` rewrites it, so edits here are lost.

## What vloop is

vloop is a command-line tool and a plugin for Claude Code and GitHub Copilot CLI. It turns a brief into tasks and, for each task, has a work session do it, a gate check it and an independent review judge it, with one commit per task. It runs on macOS, Linux and Windows.

## How to use it

Install it with `+"`go install github.com/mvelosop/vloop/cmd/vloop@latest`"+`. Run `+"`vloop doctor`"+` to check the setup. Write a brief with `+"`/vloop:create-brief`"+` in the agent and run it with `+"`vloop run <brief>`"+`. Keep vloop current with `+"`vloop self-update`"+`, then `+"`vloop upgrade`"+` in each repository.

## What is in this folder

- `+"`config.toml`"+`: shared configuration.
- `+"`config.local.toml`"+`: this machine's configuration, ignored by Git.
- `+"`install.json`"+`: vloop's installation data.
- `+"`repo-guidelines.md`"+`: repository lessons.
- `+"`README.md`"+`: this generated description.
- `+"`defects/`"+`: recorded defects.
- `+"`interventions/`"+`: recorded interventions.
- `+"`state/`"+`: the plan, runs, metrics and brief reviews.
- `+"`tmp/`"+`: scratch files, ignored by Git.

## More

https://github.com/mvelosop/vloop
`, version)
}
