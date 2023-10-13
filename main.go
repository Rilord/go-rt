package main

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"os"
	"sync"
	"time"

	"github.com/Rilord/go-rt/rendering"
	"github.com/Rilord/go-rt/types"
	"github.com/Rilord/go-rt/worker"
)

func main() {

	ctx := context.Background()

	taskPool := worker.NewTaskPool()

	pool, err := worker.NewPool(ctx, 1, "pool")

	go func() {
		pool.RunBackGround(0)
	}()
	time.Sleep(1 * time.Microsecond)

	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	aspectRatio := 16.0 / 9.0
	imgWidth := 400.0
	imgHeight := imgWidth / aspectRatio
	img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{int(imgWidth), int(imgHeight)}})

	camera := &types.Camera{
		FocalLength:    1.0,
		ViewportHeight: 2.0,
		ViewportWidth:  2.0 * (imgWidth / imgHeight),
		Center:         types.NewVec3(0, 0, 0),
	}

	viewPortUV := types.UV{
		U: types.NewVec3(camera.ViewportWidth, 0, 0),
		V: types.NewVec3(0, -camera.ViewportHeight, 0),
	}

	pixelDeltaUV := types.UV{
		U: viewPortUV.U.Scale(1 / imgWidth),
		V: viewPortUV.V.Scale(1 / imgHeight),
	}

	viewportUpperLeft := camera.Center.
		Sub(types.NewVec3(0, 0, camera.FocalLength)).
		Sub(viewPortUV.U.Scale(0.5)).
		Sub(viewPortUV.V.Scale(0.5))

	pixel00Loc := pixelDeltaUV.U.
		Add(pixelDeltaUV.V).
		Scale(0.5).
		Add(viewportUpperLeft)

	tasks := make([]*worker.Task, 0, int(imgHeight*imgWidth))

	scene := rendering.NewScene(pixel00Loc, &pixelDeltaUV, camera)

	for j := 0; j < int(imgHeight); j++ {
		for i := 0; i < int(imgWidth); i++ {

			tile := &rendering.RenderTile{X: i, Y: j}

			renderer := rendering.NewRenderer(scene, &rendering.RenderState{RenderTile: tile})

			task := worker.NewTask(ctx, taskPool, "renderPixel", nil, nil, 1, rendering.RenderPixelFn, renderer, img)
			tasks = append(tasks, task)
		}
	}

	defer func() {
		for _, task := range tasks {
			task.DeleteUnsafe(taskPool)
		}
	}()

	var wg sync.WaitGroup

	for _, task := range tasks {
		if task != nil {
			task.SetWgUnsafe(&wg)
			if err := pool.AddTask(task); err != nil {
				return
			}
		}
	}

	wg.Wait()

	for _, task := range tasks {
		if task.GetError() == nil {
			results := task.GetResults()
			for i, res := range results {
				fmt.Printf("pixel(%d): %.5f\n", i, res.(*types.Color).X)
			}
		} else {

		}
	}

	f, _ := os.Create("image.png")
	png.Encode(f, img)
}
