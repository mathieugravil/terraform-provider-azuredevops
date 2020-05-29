package azuredevops

import (
	"fmt"
	"log"
	"time"
	"context"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	//"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/core"
)

var resourceAzuredevopsProjectSchema = map[string]*schema.Schema{
	"id": {
		Type:     schema.TypeInt,
		Optional: true,
		Computed: true,
	},
	"name": {
		Type:     schema.TypeString,
		Required: true,
	},
	"description": {
		Type:     schema.TypeString,
		Required: true,
	},
	"url": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"Links": {
		Type:     schema.TypeSet,
		Optional: true,
		Computed: true,
	},
	"state": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"defaultTeam": {
		Type:     schema.TypeSet,
		Optional: true,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"name": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"url": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	},
	"revision": {
		Type:     schema.TypeInt,
		Computed: true,
	},
	"visibility": {
		Type:     schema.TypeString,
		Required: true,
		ValidateFunc: validation.StringInSlice([]string{"private",  "public"}, true),
		Default:      "private",
	},
	"lastUpdateTime": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"abbreviation": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"defaultTeamImageUrl": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"sourceControlType":{
		Type:     schema.TypeString,
		Required: true,
		ValidateFunc: validation.StringInSlice([]string{"Git",  "Tfts"}, true),
		Default:      "Git",
	},
	"templateTypeId":{
		Type:     schema.TypeString,
		Required: true,
		Default:      "adcc42ab-9882-485e-a3ed-7678f01f66bc",
	},
/*	"capabilities": {
		Type:     schema.TypeSet,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"versioncontrol": {
					Type:     schema.TypeSet,
					Required: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"sourceControlType":{
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.StringInSlice([]string{"Git",  "Tfts"}, true),
								Default:      "Git",
							},
						},
					},
				},
				"processTemplate": {
					Type:     schema.TypeSet,
					Required: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"templateTypeId":{
								Type:     schema.TypeString,
								Required: true,
								Default:      "adcc42ab-9882-485e-a3ed-7678f01f66bc",
							},
						},
					},
				},
			},
		},
	},*/

}

func resourceAzuredevopsProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceAzuredevopsProjectCreate,
		Read:   resourceAzuredevopsProjectRead,
		Update: resourceAzuredevopsProjectUpdate,
		Delete: resourceAzuredevopsProjectDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: resourceAzuredevopsProjectSchema,
	}
}

func resourceAzuredevopsProjectSetToState(d *schema.ResourceData, project *core.TeamProject) {
	var Capa = *project.Capabilities
	d.SetId(fmt.Sprintf("%d", project.Id))
	d.Set("name", project.Name)
	d.Set("description", project.Description)
	d.Set("url", project.Url)
	d.Set("links", project.Links)
	d.Set("state", project.State)
	d.Set("defaultTeam", project.DefaultTeam)
	d.Set("revision", project.Revision)
	d.Set("visibility", project.Visibility)
	d.Set("lastUpdateTime", project.LastUpdateTime)
	d.Set("abbreviation", project.Abbreviation)
	d.Set("defaultTeamImageUrl", project.DefaultTeamImageUrl)
	d.Set("sourceControlType", Capa["versioncontrol"]["sourceControlType"])
	d.Set("templateTypeId", Capa["processTemplate"]["templateTypeId"])
}

func resourceAzuredevopsProjectCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(core.Client)
	ctx := context.Background()
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	visibility :=  core.ProjectVisibility(d.Get("visibility").(string))
	var capabilities = map[string]map[string]string{}
	capabilities["versioncontrol"] =  map[string]string{}
	capabilities["versioncontrol"]["sourceControlType"] = d.Get("sourceControlType").(string)
	capabilities["processTemplate"] =  map[string]string{}
	capabilities["processTemplate"]["templateTypeId"] = d.Get("templateTypeId").(string)
	MyProject := core.TeamProject{ 
		Name : &name , 
		Description : &description , 
		Visibility :  &visibility , 
		Capabilities : &capabilities ,
	}

	QueueCreateProjectArgs := core.QueueCreateProjectArgs{
		ProjectToCreate : &MyProject ,
	}

	log.Printf("[DEBUG] create azuredevops project %q", *QueueCreateProjectArgs.ProjectToCreate.Name)

	_, err := client.QueueCreateProject(ctx, QueueCreateProjectArgs)
	if err != nil {
		return err
	}
	return resourceAzuredevopsProjectRead(d, meta)
}

func resourceAzuredevopsProjectRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(core.Client)
	ctx := context.Background()
	myprojectId := d.Get("Id").(string)
	getProjectArgs := core.GetProjectArgs{
		ProjectId : &myprojectId ,
	}
	project, _ :=  client.GetProject(ctx, getProjectArgs)
	resourceAzuredevopsProjectSetToState(d, project)
	return nil
}

func resourceAzuredevopsProjectUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceAzuredevopsProjectRead(d, meta)
}

func resourceAzuredevopsProjectDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(core.Client)
	ctx := context.Background()
	log.Printf("[DEBUG] Delete azuredevops project %s", d.Id())
	myUuid, _  := uuid.Parse(d.Id())
	queueDeleteProjectArgs := core.QueueDeleteProjectArgs{
		ProjectId : &myUuid  ,
	}
	_, err := client.QueueDeleteProject(ctx, queueDeleteProjectArgs)
	if err != nil {
		return err
	}

	// Wait for the project to be deleted.
	// Deleting a project in azuredevops is async.
	stateConf := &resource.StateChangeConf{
		Pending: []string{"Deleting"},
		Target:  []string{"Deleted"},
		Refresh: func() (interface{}, string, error) {
			myprojectId := d.Get("Id").(string)
			getProjectArgs := core.GetProjectArgs{
				ProjectId : &myprojectId  ,
			}
			response, err := client.GetProject(ctx, getProjectArgs )
			if err != nil {
				if response == nil {
					return  meta , "Deleted", nil
				}
				log.Printf("[ERROR] Received error: %#v", err)
				return  meta , "Error", err
			}
			return  meta, "Deleting", nil
		},

		Timeout:    10 * time.Minute,
		MinTimeout: 3 * time.Second,
		Delay:      5 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for project (%s) to become deleted: %s", d.Id(), err)
	}
	return nil
}

