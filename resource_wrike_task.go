package main

import (
	"fmt"
	"strconv"

	wrike "github.com/daisaru11/wrike-go"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceTask() *schema.Resource {
	return &schema.Resource{
		Create: resourceTaskCreate,
		Read:   resourceTaskRead,
		Update: resourceTaskUpdate,
		Delete: resourceTaskDelete,

		Schema: map[string]*schema.Schema{
			"title": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"status": &schema.Schema{
				Type:     schema.TypeString,
				Default:  "Active",
				Optional: true,
			},
			"importance": &schema.Schema{
				Type:     schema.TypeString,
				Default:  "Normal",
				Optional: true,
			},
			"dates": &schema.Schema{
				Type:     schema.TypeMap,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						// "duration": {
						// 	Type:     schema.TypeInt,
						// 	Optional: true,
						// },
						"start": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"due": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"work_on_weekends": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"parents": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"responsibles": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"super_tasks": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"custom_fields": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"custom_status": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func buildCreateTaskRequest(d *schema.ResourceData) (*wrike.CreateTaskRequest, error) {
	var folderID string
	if attr, ok := d.GetOk("parents"); ok {
		parentIds := attr.([]interface{})
		if len(parentIds) > 0 {
			folderID = parentIds[0].(string)
		}
	}

	if folderID == "" {
		return nil, fmt.Errorf("One or more parent folder IDs are required")
	}

	payload := wrike.CreateTaskPayload{}

	if attr, ok := d.GetOk("title"); ok {
		payload.Title = wrike.String(attr.(string))
	}

	if attr, ok := d.GetOk("description"); ok {
		payload.Description = wrike.String(attr.(string))
	}

	if attr, ok := d.GetOk("status"); ok {
		payload.Status = wrike.String(attr.(string))
	}

	if attr, ok := d.GetOk("importance"); ok {
		payload.Importance = wrike.String(attr.(string))
	}

	if attr, ok := d.GetOk("dates"); ok {
		dates := attr.(map[string]interface{})
		payload.Dates = &wrike.TaskDates{}

		if v, ok := dates["type"]; ok {
			payload.Dates.Type = wrike.String(v.(string))
		}
		// if v, ok := dates["duration"]; ok {
		// 	payload.Dates.Duration = wrike.Int(v.(int))
		// }
		if v, ok := dates["start"]; ok {
			payload.Dates.Start = wrike.String(v.(string))
		}
		if v, ok := dates["due"]; ok {
			payload.Dates.Due = wrike.String(v.(string))
		}
		if v, ok := dates["work_on_weekends"]; ok {
			payload.Dates.WorkOnWeekends = wrike.Bool(v.(bool))
		}
	}

	if attr, ok := d.GetOk("parents"); ok {
		payload.Parents = []string{}
		for _, v := range attr.([]interface{}) {
			payload.Parents = append(payload.Parents, v.(string))
		}
	}

	if attr, ok := d.GetOk("responsibles"); ok {
		payload.Responsibles = []string{}
		for _, v := range attr.([]interface{}) {
			payload.Responsibles = append(payload.Responsibles, v.(string))
		}
	}

	if attr, ok := d.GetOk("super_tasks"); ok {
		payload.SuperTasks = []string{}
		for _, v := range attr.([]interface{}) {
			payload.SuperTasks = append(payload.SuperTasks, v.(string))
		}
	}

	if attr, ok := d.GetOk("custom_fields"); ok {
		payload.CustomFields = []wrike.TaskCustomField{}

		for _, v := range attr.([]interface{}) {
			m := v.(map[string]string)
			payload.CustomFields = append(payload.CustomFields, wrike.TaskCustomField{
				ID:    wrike.String(m["id"]),
				Value: wrike.String(m["value"]),
			})
		}
	}

	if attr, ok := d.GetOk("custom_status"); ok {
		payload.CustomStatus = wrike.String(attr.(string))
	}

	return &wrike.CreateTaskRequest{
		FolderID: wrike.String(folderID),
		Payload:  &payload,
	}, nil
}

func resourceTaskCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*wrike.Client)

	req, err := buildCreateTaskRequest(d)
	if err != nil {
		return err
	}

	res, err := client.CreateTask(req)
	if err != nil {
		return fmt.Errorf("Failure on creating task: %s", err.Error())
	}

	d.SetId(wrike.StringValue(res.Data[0].ID))
	return applyTaskToResource(d, &res.Data[0])
}

func resourceTaskRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*wrike.Client)

	id := d.Id()
	req := &wrike.GetTasksRequest{
		IDs: []string{id},
	}

	res, err := client.GetTasks(req)
	if err != nil {
		return err
	}

	if len(res.Data) == 0 {
		return fmt.Errorf("Task not found. (TaskID: %s)", id)
	}

	return applyTaskToResource(d, &res.Data[0])
}

func resourceTaskUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*wrike.Client)

	req := wrike.UpdateTaskRequest{
		TaskID:  wrike.String(d.Id()),
		Payload: &wrike.UpdateTaskPayload{},
	}

	res, err := client.UpdateTask(&req)
	if err != nil {
		return fmt.Errorf("Failure on creating task: %s", err.Error())
	}

	return applyTaskToResource(d, &res.Data[0])
}

func resourceTaskDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func applyTaskToResource(d *schema.ResourceData, task *wrike.Task) error {
	d.Set("title", wrike.StringValue(task.Title))

	if task.Description != nil {
		d.Set("description", wrike.StringValue(task.Description))
	}
	if task.Status != nil {
		d.Set("status", wrike.StringValue(task.Status))
	}
	if task.Importance != nil {
		d.Set("importance", wrike.StringValue(task.Importance))
	}
	if task.Dates != nil {
		dates := make(map[string]string)

		dates["type"] = wrike.StringValue(task.Dates.Type)
		// if task.Dates.Duration != nil {
		// 	dates["duration"] = strconv.Itoa(wrike.IntValue(task.Dates.Duration))
		// }
		if task.Dates.Start != nil {
			dates["start"] = wrike.StringValue(task.Dates.Start)
		}
		if task.Dates.Due != nil {
			dates["due"] = wrike.StringValue(task.Dates.Due)
		}
		if task.Dates.WorkOnWeekends != nil {
			dates["work_on_weekends"] = strconv.FormatBool(wrike.BoolValue(task.Dates.WorkOnWeekends))
		}

		d.Set("dates", dates)
	}

	if task.ParentIDs != nil {
		d.Set("parents", task.ParentIDs)
	}
	if task.ResponsibleIDs != nil {
		d.Set("responsibles", task.ResponsibleIDs)
	}
	if task.SuperTaskIDs != nil {
		d.Set("super_tasks", task.SuperTaskIDs)
	}
	if task.CustomFields != nil {
		customFields := []map[string]string{}
		for _, f := range task.CustomFields {
			field := make(map[string]string)
			field["id"] = wrike.StringValue(f.ID)
			field["value"] = wrike.StringValue(f.Value)

			customFields = append(customFields, field)
		}
		d.Set("custom_fields", task.SuperTaskIDs)
	}
	if task.CustomStatusID != nil {
		d.Set("custom_status", wrike.StringValue(task.CustomStatusID))
	}

	return nil
}
