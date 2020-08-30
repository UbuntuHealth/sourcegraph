// Code generated by stringdata. DO NOT EDIT.

package schema

// CampaignSpecSchemaJSON is the content of the file "campaign_spec.schema.json".
const CampaignSpecSchemaJSON = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "CampaignSpec",
  "description": "A campaign specification, which describes the campaign and what kinds of changes to make (or what existing changesets to track).",
  "type": "object",
  "additionalProperties": false,
  "required": ["name"],
  "properties": {
    "name": {
      "type": "string",
      "description": "The name of the campaign, which is unique among all campaigns in the namespace. A campaign's name is case-preserving.",
      "pattern": "^[\\w.-]+$"
    },
    "description": {
      "type": "string",
      "description": "The description of the campaign."
    },
    "on": {
      "type": "array",
      "description": "The set of repositories (and branches) to run the campaign on, specified as a list of search queries (that match repositories) and/or specific repositories.",
      "items": {
        "title": "OnQueryOrRepository",
        "oneOf": [
          {
            "title": "OnQuery",
            "type": "object",
            "description": "A Sourcegraph search query that matches a set of repositories (and branches). Each matched repository branch is added to the list of repositories that the campaign will be run on.",
            "additionalProperties": false,
            "required": ["repositoriesMatchingQuery"],
            "properties": {
              "repositoriesMatchingQuery": {
                "type": "string",
                "description": "A Sourcegraph search query that matches a set of repositories (and branches). If the query matches files, symbols, or some other object inside a repository, the object's repository is included.",
                "examples": ["file:README.md"]
              }
            }
          },
          {
            "title": "OnRepository",
            "type": "object",
            "description": "A specific repository (and branch) that is added to the list of repositories that the campaign will be run on.",
            "additionalProperties": false,
            "required": ["repository"],
            "properties": {
              "repository": {
                "type": "string",
                "description": "The name of the repository (as it is known to Sourcegraph).",
                "examples": ["github.com/foo/bar"]
              },
              "branch": {
                "type": "string",
                "description": "The branch on the repository to propose changes to. If unset, the repository's default branch is used."
              }
            }
          }
        ]
      }
    },
    "steps": {
      "type": "array",
      "description": "The sequence of commands to run (for each repository branch matched in the ` + "`" + `on` + "`" + ` property) to produce the campaign's changes.",
      "items": {
        "title": "Step",
        "type": "object",
        "description": "A command to run (as part of a sequence) in a repository branch to produce the campaign's changes.",
        "additionalProperties": false,
        "required": ["run", "container"],
        "properties": {
          "run": {
            "type": "string",
            "description": "The shell command to run in the container. It can also be a multi-line shell script. The working directory is the root directory of the repository checkout."
          },
          "container": {
            "type": "string",
            "description": "The Docker image used to launch the Docker container in which the shell command is run.",
            "examples": ["alpine:3"]
          },
          "env": {
            "type": "object",
            "description": "Environment variables to set in the environment when running this command.",
            "additionalProperties": {
              "type": "string"
            }
          }
        }
      }
    },
    "importChangesets": {
      "type": "array",
      "description": "Import existing changesets on code hosts.",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["repository", "externalIDs"],
        "properties": {
          "repository": {
            "type": "string",
            "description": "The repository name as configured on your Sourcegraph instance."
          },
          "externalIDs": {
            "type": "array",
            "description": "The changesets to import from the code host. For GitHub this is the PR number, for GitLab this is the MR number, for Bitbucket Server this is the PR number.",
            "uniqueItems": true,
            "items": {
              "oneOf": [{ "type": "string" }, { "type": "integer" }]
            },
            "examples": [120, "120"]
          }
        }
      }
    },
    "changesetTemplate": {
      "type": "object",
      "description": "A template describing how to create (and update) changesets with the file changes produced by the command steps.",
      "additionalProperties": false,
      "required": ["title", "branch", "commit", "published"],
      "properties": {
        "title": {
          "description": "The title of the changeset.",
          "oneOf": [
            {
              "type": "string",
              "description": "The title to use for the entire campaign."
            },
            {
              "type": "object",
              "required": ["default", "only"],
              "additionalProperties": false,
              "properties": {
                "default": {
                  "type": "string",
                  "description": "The title to use for all changesets that do not match any of the rules in the only array."
                },
                "only": {
                  "type": "array",
                  "items": {
                    "type": "object",
                    "required": ["match", "value"],
                    "additionalProperties": false,
                    "properties": {
                      "match": {
                        "type": "string",
                        "description": "The repository name to match. Glob wildcards are supported."
                      },
                      "value": {
                        "type": "string",
                        "description": "The title to use for changesets that match this rule."
                      }
                    }
                  }
                }
              }
            }
          ]
        },
        "body": {
          "type": "string",
          "description": "The body (description) of the changeset."
        },
        "branch": {
          "type": "string",
          "description": "The name of the Git branch to create or update on each repository with the changes."
        },
        "commit": {
          "title": "ExpandedGitCommitDescription",
          "type": "object",
          "description": "The Git commit to create with the changes.",
          "additionalProperties": false,
          "required": ["message"],
          "properties": {
            "message": {
              "type": "string",
              "description": "The Git commit message."
            }
          }
        },
        "published": {
          "description": "Whether to publish the changeset. An unpublished changeset can be previewed on Sourcegraph by any person who can view the campaign, but its commit, branch, and pull request aren't created on the code host. A published changeset results in a commit, branch, and pull request being created on the code host.",
          "oneOf": [
            {
              "type": "boolean",
              "description": "A single flag to control the publishing state for the entire campaign."
            },
            {
              "type": "object",
              "title": "PublishedOnly",
              "description": "Only repositories that match patterns in this array will be published.",
              "additionalProperties": false,
              "required": ["only"],
              "properties": {
                "only": {
                  "type": "array",
                  "items": {
                    "type": "string"
                  }
                }
              }
            },
            {
              "type": "object",
              "title": "PublishedExcept",
              "description": "Only repositories that do NOT match patterns in this array will be published.",
              "additionalProperties": false,
              "required": ["except"],
              "properties": {
                "except": {
                  "type": "array",
                  "items": {
                    "type": "string"
                  }
                }
              }
            }
          ]
        }
      }
    }
  }
}
`
