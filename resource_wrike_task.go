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
			"title": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"status": {
				Type:     schema.TypeString,
				Default:  "Active",
				Optional: true,
			},
			"importance": {
				Type:     schema.TypeString,
				Default:  "Normal",
				Optional: true,
			},
			"dates": {
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
			"parents": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"responsibles": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"super_tasks": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"custom_fields": {
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
			"custom_status": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func buildCreateTaskRequest(d *schema.ResourceData) (*wrike.CreateTaskRequest, error) {
	var folderID string
	if attr, ok := d.GetOk("parents"); ok {
		parentIds := expandStringSet(attr.(*schema.Set))
		if len(parentIds) > 0 {
			folderID = parentIds[0]
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
		payload.Parents = expandStringSet(attr.(*schema.Set))
	}

	if attr, ok := d.GetOk("responsibles"); ok {
		payload.Responsibles = expandStringSet(attr.(*schema.Set))
	}

	if attr, ok := d.GetOk("super_tasks"); ok {
		payload.SuperTasks = expandStringSet(attr.(*schema.Set))
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
		return fmt.Errorf("Failure on creating task: %s", err.Error())
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

//nolint:unparam
func buildUpdateTaskRequest(d *schema.ResourceData) (*wrike.UpdateTaskRequest, error) {
	payload := wrike.UpdateTaskPayload{}

	if d.HasChange("title") {
		_, attr := d.GetChange("title")
		payload.Title = wrike.String(attr.(string))
	}

	if d.HasChange("description") {
		_, attr := d.GetChange("description")
		payload.Description = wrike.String(attr.(string))
	}

	if d.HasChange("status") {
		_, attr := d.GetChange("status")
		payload.Status = wrike.String(attr.(string))
	}

	if d.HasChange("importance") {
		_, attr := d.GetChange("importance")
		payload.Importance = wrike.String(attr.(string))
	}

	if d.HasChange("dates") {
		attr := d.Get("dates")
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

	if d.HasChange("parents") {
		old, new := d.GetChange("parents")
		oldSet := old.(*schema.Set)
		newSet := new.(*schema.Set)

		if added := newSet.Difference(oldSet); added.Len() > 0 {
			payload.AddParents = expandStringSet(added)
		}

		if removed := oldSet.Difference(newSet); removed.Len() > 0 {
			payload.RemoveParents = expandStringSet(removed)
		}
	}

	if d.HasChange("responsibles") {
		old, new := d.GetChange("responsibles")
		oldSet := old.(*schema.Set)
		newSet := new.(*schema.Set)

		if added := newSet.Difference(oldSet); added.Len() > 0 {
			payload.AddResponsibles = expandStringSet(added)
		}

		if removed := oldSet.Difference(newSet); removed.Len() > 0 {
			payload.RemoveResponsibles = expandStringSet(removed)
		}
	}

	if d.HasChange("super_tasks") {
		old, new := d.GetChange("super_tasks")
		oldSet := old.(*schema.Set)
		newSet := new.(*schema.Set)

		if added := newSet.Difference(oldSet); added.Len() > 0 {
			payload.AddSuperTasks = expandStringSet(added)
		}

		if removed := oldSet.Difference(newSet); removed.Len() > 0 {
			payload.RemoveSuperTasks = expandStringSet(removed)
		}
	}

	return &wrike.UpdateTaskRequest{
		TaskID:  wrike.String(d.Id()),
		Payload: &payload,
	}, nil
}

func resourceTaskUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*wrike.Client)

	req, err := buildUpdateTaskRequest(d)
	if err != nil {
		return fmt.Errorf("Failure on updating task: %s", err.Error())
	}

	res, err := client.UpdateTask(req)
	if err != nil {
		return fmt.Errorf("Failure on updating task: %s", err.Error())
	}

	return applyTaskToResource(d, &res.Data[0])
}

func resourceTaskDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*wrike.Client)

	id := d.Id()
	req := &wrike.DeleteTaskRequest{
		TaskID: wrike.String(id),
	}
	if _, err := client.DeleteTask(req); err != nil {
		return fmt.Errorf("Failure on deleting task: %s", err.Error())
	}

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
		d.Set("custom_fields", customFields)
	}
	if task.CustomStatusID != nil {
		d.Set("custom_status", wrike.StringValue(task.CustomStatusID))
	}

	return nil
}
