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
	"name": {
		Type:     schema.TypeString,
		Required: true,
	},
	"description": {
		Type:     schema.TypeString,
		Required: true,
	},
	"visibility": {
		Type:     schema.TypeString,
		Optional: true,
		ValidateFunc: validation.StringInSlice([]string{"private",  "public"}, true),
		Default:      "private",
	},
	"source_control_type":{
		Type:     schema.TypeString,
		Optional: true,
		ValidateFunc: validation.StringInSlice([]string{"git",  "tfts"}, true),
		Default:      "git",
	},
	"template_type_id":{
		Type:     schema.TypeString,
		Optional: true,
		Default:      "adcc42ab-9882-485e-a3ed-7678f01f66bc",
	},

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
	d.SetId((project.Id).String())
	d.Set("name", project.Name)
	d.Set("description", project.Description)
	d.Set("visibility", project.Visibility)
	d.Set("sourceControlType", Capa["versioncontrol"]["source_control_type"])
	d.Set("templateTypeId", Capa["processTemplate"]["template_type_id"])
}

func resourceAzuredevopsProjectCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*core.Client)
	myclient := *client
	ctx := context.Background()
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	visibility :=  core.ProjectVisibility(d.Get("visibility").(string))
	var capabilities = map[string]map[string]string{}
	capabilities["versioncontrol"] =  map[string]string{}
	capabilities["versioncontrol"]["sourceControlType"] = d.Get("source_control_type").(string)
	capabilities["processTemplate"] =  map[string]string{}
	capabilities["processTemplate"]["templateTypeId"] = d.Get("template_type_id").(string)
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

	_, err := myclient.QueueCreateProject(ctx, QueueCreateProjectArgs)
	if err != nil {
		return err
	}
	
	time.Sleep(10 * time.Second)
	// Get first page of the list of team projects for your organization
	top := 3000
	responseValue, err := myclient.GetProjects(ctx, core.GetProjectsArgs{Top: &top})
	if err != nil {
		log.Fatal(err)
	}
	for responseValue != nil {
		// Log the page of team project names
		for _, teamProjectReference := range (*responseValue).Value {
			log.Printf(" %v == %v ? ", name, *teamProjectReference.Name)
			if *teamProjectReference.Name == name {
			log.Printf("Name[%v] = %v", *teamProjectReference.Id, *teamProjectReference.Name)
			myprojectId := (*teamProjectReference.Id).String()
			d.SetId(myprojectId )
			log.Printf(" BREAK ")
			break
			}
		}
		// if continuationToken has a value, then there is at least one more page of projects to get
		if responseValue.ContinuationToken != "" {
			// Get next page of team projects
			projectArgs := core.GetProjectsArgs{
				ContinuationToken: &responseValue.ContinuationToken,
			}
			responseValue, err = myclient.GetProjects(ctx, projectArgs)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			responseValue = nil
		}
	}
	log.Printf("[DEBUG] BEFORE Read azuredevops project %q", d.Get("name").(string))
	return resourceAzuredevopsProjectRead(d, meta)
}

func resourceAzuredevopsProjectRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*core.Client)
	myclient := *client
	ctx := context.Background()
	myprojectId := d.Get("Id").(string)
	log.Printf("[DEBUG] Read azuredevops project %q", d.Get("name").(string))
	getProjectArgs := core.GetProjectArgs{
		ProjectId : &myprojectId ,
	}
	project, _ :=  myclient.GetProject(ctx, getProjectArgs)
	resourceAzuredevopsProjectSetToState(d, project)
	return nil
}

func resourceAzuredevopsProjectUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceAzuredevopsProjectRead(d, meta)
}

func resourceAzuredevopsProjectDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*core.Client)
	myclient := *client
	ctx := context.Background()
	log.Printf("[DEBUG] Delete azuredevops project %s", d.Id())
	myUuid, _  := uuid.Parse(d.Id())
	queueDeleteProjectArgs := core.QueueDeleteProjectArgs{
		ProjectId : &myUuid  ,
	}
	_, err := myclient.QueueDeleteProject(ctx, queueDeleteProjectArgs)
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
			response, err := myclient.GetProject(ctx, getProjectArgs )
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

